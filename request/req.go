package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/gohandy/marshaling"
	"github.com/materials-commons/mcfs/protocol"
	"io"
	"reflect"
)

const maxBadRequests = 10

type reqStateFN func() reqStateFN

type db struct {
	session *r.Session
}

// ReqHandler is an instance of the request state machine for handling client requests.
type ReqHandler struct {
	session *r.Session
	user    string
	mcdir   string
	marshaling.MarshalUnmarshaler
	badRequestCount int
}

type stateStatus struct {
	status mc.ErrorCode
	err    error
}

func ss(statusCode mc.ErrorCode, err error) *stateStatus {
	return &stateStatus{
		status: statusCode,
		err:    err,
	}
}

func ssf(statusCode mc.ErrorCode, message string, args ...interface{}) *stateStatus {
	err := fmt.Errorf(message, args...)
	return &stateStatus{
		status: statusCode,
		err:    err,
	}
}

// NewReqHandler creates a new ReqHandlerInstance. Each ReqHandler is a thread safe state machine for
// handling client requests.
func NewReqHandler(m marshaling.MarshalUnmarshaler, session *r.Session, mcdir string) *ReqHandler {
	return &ReqHandler{
		session:            session,
		MarshalUnmarshaler: m,
		mcdir:              mcdir,
	}
}

// Run run the ReqHandler state machine. The state machine accepts and processes request according
// to the mcfs.protocol package.
func (h *ReqHandler) Run() {
	for reqStateFN := h.startState; reqStateFN != nil; {
		reqStateFN = reqStateFN()
	}
}

type errorReq struct{}

func (h *ReqHandler) req() interface{} {
	var req protocol.Request
	if err := h.Unmarshal(&req); err != nil {
		if err == io.EOF {
			return protocol.CloseReq{}
		}
		return errorReq{}
	}
	return req.Req
}

func (h *ReqHandler) startState() reqStateFN {
	var resp interface{}
	var s *stateStatus
	request := h.req()
	switch req := request.(type) {
	case protocol.LoginReq:
		resp, s = h.login(&req)
		if s != nil {
			return h.badRequestRestart(s)
		}
		h.respOk(resp)
		return h.nextCommand
	case protocol.CloseReq:
		return nil
	default:
		return h.badRequestRestart(ssf(mc.ErrorCodeInvalid, "Bad Request %T", req))
	}
}

func (h *ReqHandler) badRequestRestart(s *stateStatus) reqStateFN {
	fmt.Println("badRequestRestart:", s.status, s.err)
	h.badRequestCount = h.badRequestCount + 1
	resp := &protocol.Response{
		Status:        s.status,
		StatusMessage: s.err.Error(),
	}
	h.Marshal(resp)
	if h.badRequestCount > maxBadRequests {
		return nil
	}
	return h.startState
}

func (h *ReqHandler) badRequestNext(s *stateStatus) reqStateFN {
	fmt.Println("badRequestNext:", s.status, s.err)
	resp := &protocol.Response{
		Status:        s.status,
		StatusMessage: s.err.Error(),
	}
	h.Marshal(resp)
	if h.badRequestCount > maxBadRequests {
		return nil
	}
	return h.nextCommand
}

func (h *ReqHandler) nextCommand() reqStateFN {
	var s *stateStatus
	var resp interface{}

	request := h.req()
	switch req := request.(type) {
	case protocol.UploadReq:
		var respUpload *protocol.UploadResp
		respUpload, s = h.upload(&req)
		if s == nil {
			return h.uploadLoop(respUpload)
		}
	case protocol.CreateFileReq:
		resp, s = h.createFile(&req)
	case protocol.CreateDirReq:
		resp, s = h.createDir(&req)
	case protocol.CreateProjectReq:
		resp, s = h.createProject(&req)
	case protocol.DownloadReq:
	case protocol.MoveReq:
	case protocol.DeleteReq:
	case protocol.StatProjectReq:
		resp, s = h.statProject(&req)
	case protocol.LookupReq:
		resp, s = h.lookup(&req)
	case protocol.LogoutReq:
		resp, s = h.logout(&req)
		h.sendResp(resp, s)
		return h.startState
	case protocol.StatReq:
		resp, s = h.stat(&req)
	case protocol.CloseReq:
		return nil
	case protocol.IndexReq:
	default:
		h.badRequestCount = h.badRequestCount + 1
		return h.badRequestNext(ssf(mc.ErrorCodeInvalid, "Bad request %T", req))
	}

	h.sendResp(resp, s)
	return h.nextCommand
}

func (h *ReqHandler) sendResp(resp interface{}, s *stateStatus) {
	if s != nil {
		h.respError(resp, s)
	} else {
		h.respOk(resp)
	}
}

func (h *ReqHandler) respOk(respData interface{}) {
	resp := &protocol.Response{
		Status: mc.ErrorCodeSuccess,
		Resp:   respData,
	}
	err := h.Marshal(resp)
	if err != nil {
		fmt.Println("respOk, marshal error = ", err)
	}
}

func (h *ReqHandler) respError(respData interface{}, s *stateStatus) {
	fmt.Println("respError =", s.status, ",", s.err)
	resp := &protocol.Response{
		Status:        s.status,
		StatusMessage: s.err.Error(),
	}

	if !reflect.ValueOf(respData).IsNil() {
		resp.Resp = respData
	}

	err := h.Marshal(resp)
	if err != nil {
		fmt.Println("respError, marshal error = ", err)
	}
}
