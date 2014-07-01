package request

import (
	"fmt"
	"io"

	"github.com/materials-commons/mcfs/base/codex"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/protocol"
	"github.com/materials-commons/mcfs/server/inuse"
	"github.com/materials-commons/mcfs/server/service"
)

const maxBadRequests = 10

type reqStateFN func() reqStateFN

type errorReq struct{}

type ReqHandler struct {
	codex           *protocol.Codex
	badRequestCount int
	user            string
	mcdir           string
	projectID       string
	buf             []byte
	service         *service.Service
	io.ReadWriter
}

const maxBufSize = (1024 * 1024 * 20) + 2048

func NewReqHandler(rw io.ReadWriter, encoderDecoder codex.EncoderDecoder, mcdir string) *ReqHandler {
	return &ReqHandler{
		ReadWriter: rw,
		codex:      protocol.NewCodex(encoderDecoder),
		mcdir:      mcdir,
		buf:        make([]byte, maxBufSize),
		service:    service.New(service.RethinkDB),
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

func (h *ReqHandler) req() interface{} {
	_, err := h.Read(h.buf)
	if err != nil {
		switch {
		case err == io.EOF:
			return protocol.CloseReq{}
		default:
			return &errorReq{}
		}
	}

	request, err := h.codex.Decode(h.buf)
	if err != nil {
		return &errorReq{}
	}

	return request
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
		return h.badRequestRestart(mcerr.Errorf(mcerr.ErrInvalid, "Bad Request %T", req))
	}
}

func (h *ReqHandler) badRequestRestart(err error) reqStateFN {
	fmt.Println("badRequestRestart:", err)

	h.respError(err)
	h.badRequestCount = h.badRequestCount + 1
	if h.badRequestCount > maxBadRequests {
		return nil
	}
	return h.startState
}

func (h *ReqHandler) badRequestNext(err error) reqStateFN {
	fmt.Println("badRequestNext:", err)
	h.respError(err)
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
	case protocol.CreateDirectoryReq:
		resp, err = h.createDir(&req)
	case protocol.CreateProjectReq:
		resp, err = h.createProject(&req)
		//	case protocol.StatProjectReq:
		//		resp, err = h.statProject(&req)
		//	case protocol.LookupReq:
		//		resp, err = h.lookup(&req)
	case protocol.LogoutReq:
		_ = h.logout(&req)
		return h.startState
		//	case protocol.StatReq:
		//		resp, err = h.stat(&req)
	case protocol.CloseReq:
		return nil
	default:
		h.badRequestCount = h.badRequestCount + 1
		return h.badRequestNext(mcerr.Errorf(mcerr.ErrInvalid, "Bad request %T", req))
	}

	h.sendResp(resp, err)
	return h.nextCommand
}

func (h *ReqHandler) sendResp(resp interface{}, err error) {
	if err != nil {
		h.respError(err)
	} else {
		h.respOk(resp)
	}
}

func (h *ReqHandler) respOk(resp interface{}) {
	b, err := h.codex.Encode(resp, 0)
	if err != nil {
		fmt.Println("respOk: marshal error = ", err)
	}
	// How to send on?
	b.WriteTo(h)
}

func (h *ReqHandler) respError(err error) {
	var resp protocol.ErrorResp
	switch e := err.(type) {
	case *mcerr.Error:
		resp.Err = e.ToErrorCode()
		resp.Message = e.Error()
	default:
		resp.Err = mcerr.ErrorToErrorCode(err)
	}

	b, encodeErr := h.codex.Encode(resp, 0)
	if encodeErr != nil {
		fmt.Println("respError: encode error = ", encodeErr)
	}
	b.WriteTo(h)
}
