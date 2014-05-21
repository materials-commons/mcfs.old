package request

import (
	"fmt"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/gohandy/marshaling"
	"github.com/materials-commons/mcfs/inuse"
	"github.com/materials-commons/mcfs/protocol"
	"io"
	"reflect"
)

const maxBadRequests = 10

type reqStateFN func() reqStateFN

// ReqHandler is an instance of the request state machine for handling client requests.
type ReqHandler struct {
	user            string // User who connected
	projectID       string // The project that is being uploaded
	mcdir           string // Location of the materials commons data directory
	badRequestCount int    // Keep track of bad requests. Close connection when too many.
	marshaling.MarshalUnmarshaler
}

// NewReqHandler creates a new ReqHandlerInstance. Each ReqHandler is a thread safe state machine for
// handling client requests.
func NewReqHandler(m marshaling.MarshalUnmarshaler, mcdir string) *ReqHandler {
	return &ReqHandler{
		MarshalUnmarshaler: m,
		mcdir:              mcdir,
	}
}

// Run run the ReqHandler state machine. It also performs any needed cleanup when
// the state machine finishes. The state machine accepts and processes request
// according to the mcfs.protocol package.
func (h *ReqHandler) Run() {
	for reqStateFN := h.startState; reqStateFN != nil; {
		reqStateFN = reqStateFN()
	}

	// Release project lock.
	inuse.Unmark(h.projectID)
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
	var err error
	request := h.req()
	switch req := request.(type) {
	case protocol.LoginReq:
		resp, err = h.login(&req)
		if err != nil {
			return h.badRequestRestart(err)
		}
		h.respOk(resp)
		return h.nextCommand
	case protocol.CloseReq:
		return nil
	default:
		return h.badRequestRestart(mc.Errorf(mc.ErrInvalid, "Bad Request %T", req))
	}
}

func (h *ReqHandler) badRequestRestart(err error) reqStateFN {
	fmt.Println("badRequestRestart:", err)

	// Need to pass a fake response to respError that is nil.
	var resp *protocol.LoginResp
	h.respError(resp, err)
	h.badRequestCount = h.badRequestCount + 1
	if h.badRequestCount > maxBadRequests {
		return nil
	}
	return h.startState
}

func (h *ReqHandler) badRequestNext(err error) reqStateFN {
	fmt.Println("badRequestNext:", err)
	h.respError(nil, err)
	if h.badRequestCount > maxBadRequests {
		return nil
	}
	return h.nextCommand
}

func (h *ReqHandler) nextCommand() reqStateFN {
	var err error
	var resp interface{}

	request := h.req()
	switch req := request.(type) {
	case protocol.UploadReq:
		var respUpload *protocol.UploadResp
		respUpload, err = h.upload(&req)
		if err == nil {
			return h.uploadLoop(respUpload)
		}
	case protocol.CreateFileReq:
		resp, err = h.createFile(&req)
	case protocol.CreateDirReq:
		resp, err = h.createDir(&req)
	case protocol.CreateProjectReq:
		resp, err = h.createProject(&req)
	case protocol.DownloadReq:
	case protocol.MoveReq:
	case protocol.DeleteReq:
	case protocol.StatProjectReq:
		resp, err = h.statProject(&req)
	case protocol.LookupReq:
		resp, err = h.lookup(&req)
	case protocol.LogoutReq:
		resp, err = h.logout(&req)
		h.sendResp(resp, err)
		return h.startState
	case protocol.StatReq:
		resp, err = h.stat(&req)
	case protocol.CloseReq:
		return nil
	case protocol.IndexReq:
	default:
		h.badRequestCount = h.badRequestCount + 1
		return h.badRequestNext(mc.Errorf(mc.ErrInvalid, "Bad request %T", req))
	}

	h.sendResp(resp, err)
	return h.nextCommand
}

func (h *ReqHandler) sendResp(resp interface{}, err error) {
	if err != nil {
		h.respError(resp, err)
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

func (h *ReqHandler) respError(respData interface{}, err error) {
	var resp protocol.Response
	switch e := err.(type) {
	case *mc.Error:
		resp.Status = e.ToErrorCode()
		resp.StatusMessage = e.Error()
	default:
		resp.Status = mc.ErrorToErrorCode(err)
	}

	if !reflect.ValueOf(respData).IsNil() {
		resp.Resp = respData
	}

	marshalErr := h.Marshal(resp)
	if marshalErr != nil {
		fmt.Println("respError, marshal error = ", marshalErr)
	}
}
