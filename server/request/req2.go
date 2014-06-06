package request

import (
	"github.com/materials-commons/gohandy/marshaling"
	"github.com/materials-commons/mcfs/protocol"
	"io"
)

// ReqHandler is an instance of the request state machine for handling client requests.
type oReqHandler struct {
	user            string // User who connected
	projectID       string // The project that is being uploaded
	mcdir           string // Location of the materials commons data directory
	badRequestCount int    // Keep track of bad requests. Close connection when too many.
	marshaling.MarshalUnmarshaler
}

// NewReqHandler creates a new ReqHandlerInstance. Each ReqHandler is a thread safe state machine for
// handling client requests.
func NewoReqHandler(m marshaling.MarshalUnmarshaler, mcdir string) *oReqHandler {
	return &oReqHandler{
		MarshalUnmarshaler: m,
		mcdir:              mcdir,
	}
}

func (h *oReqHandler) req() interface{} {
	var req protocol.Request
	if err := h.Unmarshal(&req); err != nil {
		if err == io.EOF {
			return protocol.CloseReq{}
		}
		return errorReq{}
	}
	return req.Req
}
