package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/codex"
	"github.com/materials-commons/base/mc"
	//"github.com/materials-commons/gohandy/marshaling"
	p2 "github.com/materials-commons/base/protocol"
	"github.com/materials-commons/mcfs/protocol"
	"io"
	"reflect"
)

const maxBadRequests = 10

const maxBufSize = (1024 * 1024 * 20) + (1024 + 1024)

type reqStateFN func() reqStateFN

// ReqHandler is an instance of the request state machine for handling client requests.
type ReqHandler struct {
	session         *r.Session
	user            string
	mcdir           string
	badRequestCount int
	buf             []byte
	io.ReadWriter
	codex.EncoderDecoder
}

// NewReqHandler creates a new ReqHandlerInstance. Each ReqHandler is a thread safe state machine for
// handling client requests.
func NewReqHandler(rw io.ReadWriter, encoderDecoder codex.EncoderDecoder, session *r.Session, mcdir string) *ReqHandler {
	return &ReqHandler{
		session:        session,
		ReadWriter:     rw,
		EncoderDecoder: encoderDecoder,
		mcdir:          mcdir,
		buf:            make([]byte, maxBufSize),
	}
}

// Run run the ReqHandler state machine. The state machine accepts and processes request according
// to the mcfs.protocol package.
func (h *ReqHandler) Run() {
	for reqStateFN := h.startState; reqStateFN != nil; {
		reqStateFN = reqStateFN()
	}
}

func (h *ReqHandler) req() interface{} {
	bytesReader, err := h.Read(h.buf)
	if err != nil {
		if err == io.EOF {
			return protocol.CloseReq{}
		}
		return errorReq{}
	}

	pb, err := h.Prepare(h.buf)
	if err != nil {
		return errorReq{}
	}

	var req interface{}
	switch pb.Type {
	case p2.LoginRequest:
		var lr p2.LoginReq
		err = h.Decode(pb.Bytes, &lr)
		req = &lr
	case p2.LogoutRequest:
		var lr p2.LogoutReq
		err = h.Decode(pb.Bytes, &lr)
		req = &lr
	case p2.CreateProjectRequest:
		var cpr p2.CreateProjectReq
		err = h.Decode(pb.Bytes, &cpr)
		req = &cpr
	case p2.CreateDirectoryRequest:
		var cdr p2.CreateDirectoryReq
		err = h.Decode(pb.Bytes, &cdr)
		req = &cdr
	case p2.CreateFileRequest:
		var cfr p2.CreateFileReq
		err = h.Decode(pb.Bytes, &cfr)
		req = &cfr
	case p2.DirectoryStatRequest:
		var dsr p2.DirectoryStatReq
		err = h.Decode(pb.Bytes, &dsr)
		req = &dsr
	}

	if err != nil {
		return errorReq{}
	}

	return req
}

type errorReq struct{}

/*
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
*/

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
	h.badRequestCount = h.badRequestCount + 1
	resp := &protocol.Response{
		Status:        mc.ErrorToErrorCode(err),
		StatusMessage: err.Error(),
	}
	h.Marshal(resp)
	if h.badRequestCount > maxBadRequests {
		return nil
	}
	return h.startState
}

func (h *ReqHandler) badRequestNext(err error) reqStateFN {
	fmt.Println("badRequestNext:", err)
	resp := &protocol.Response{
		Status:        mc.ErrorToErrorCode(err),
		StatusMessage: err.Error(),
	}
	h.Marshal(resp)
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
	var resp *protocol.Response

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
