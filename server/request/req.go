package request

import (
	"fmt"
	"io"
	"reflect"

	"github.com/materials-commons/gohandy/marshaling"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/server/inuse"
	"github.com/materials-commons/mcfs/server/service"
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
	service *service.Service
}

// NewReqHandler creates a new ReqHandlerInstance. Each ReqHandler is a thread safe state machine for
// handling client requests.
func NewReqHandler(m marshaling.MarshalUnmarshaler, mcdir string) *ReqHandler {
	return &ReqHandler{
		MarshalUnmarshaler: m,
		mcdir:              mcdir,
		service:            service.New(service.RethinkDB),
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
		return h.badRequestRestart(mcerr.Errorf(mcerr.ErrInvalid, "Bad Request %T", req))
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
		fmt.Println("UploadReq")
		var respUpload *protocol.UploadResp
		fmt.Println("upload")
		respUpload, err = h.upload(&req)
		fmt.Println("past upload")
		if err == nil {
			fmt.Println("uploadLoop")
			return h.uploadLoop(respUpload)
			fmt.Println("past uploadLoop")
		}
	case protocol.CreateFileReq:
		fmt.Println("CreateFileReq")
		resp, err = h.createFile(&req)
		fmt.Println("left createFile")
	case protocol.CreateDirReq:
		fmt.Println("CreateDirReq")
		resp, err = h.createDir(&req)
	case protocol.CreateProjectReq:
		fmt.Println("CreateProjectReq")
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
		fmt.Println("bad request")
		h.badRequestCount = h.badRequestCount + 1
		return h.badRequestNext(mcerr.Errorf(mcerr.ErrInvalid, "Bad request %T", req))
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
		Status: mcerr.ErrorCodeSuccess,
		Resp:   respData,
	}
	err := h.Marshal(resp)
	if err != nil {
		fmt.Println("respOk: marshal error = ", err)
	}
}

func (h *ReqHandler) respError(respData interface{}, err error) {
	var resp protocol.Response
	switch e := err.(type) {
	case *mcerr.Error:
		resp.Status = e.ToErrorCode()
		resp.StatusMessage = e.Error()
	default:
		resp.Status = mcerr.ErrorToErrorCode(err)
	}

	fmt.Println("respError: ", resp.Status, resp.StatusMessage)

	if respData != nil && !reflect.ValueOf(respData).IsNil() {
		resp.Resp = respData
	}

	marshalErr := h.Marshal(resp)
	if marshalErr != nil {
		fmt.Println("respError: marshal error = ", marshalErr)
	}
}
