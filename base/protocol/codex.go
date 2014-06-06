package protocol

import (
	"bytes"
	"errors"
	"github.com/materials-commons/mcfs/base/codex"
)

// ErrBadType The message type is unknown
var ErrBadType = errors.New("bad type")

// A Codex will encode and decode messages.
type Codex struct {
	codex codex.EncoderDecoder
}

// NewCodex creates a new Codex instance.
func NewCodex(encoderDecoder codex.EncoderDecoder) *Codex {
	return &Codex{codex: encoderDecoder}
}

// Decode will decode bytes to the known type.
func (c *Codex) Decode(b []byte) (interface{}, error) {
	pb, err := c.codex.Prepare(b)
	if err != nil {
		return nil, err
	}

	var item interface{}

	switch pb.Type {
	case LoginRequest:
		var req LoginReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case LogoutRequest:
		var req LogoutReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case CreateProjectRequest:
		var req CreateProjectReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case CreateDirectoryRequest:
		var req CreateDirectoryReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case CreateFileRequest:
		var req CreateFileReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case DirectoryStatRequest:
		var req DirectoryStatReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case UploadRequest:
		var req UploadReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case UploadBytesRequest:
		var req UploadBytesReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case DoneRequest:
		var req DoneReq
		err = c.codex.Decode(pb.Bytes, &req)
		if err != nil {
			item = &req
		}

	case LoginResponse:
		var resp LoginResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	case CreateProjectResponse:
		var resp CreateProjectResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	case CreateDirectoryResponse:
		var resp CreateDirectoryResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	case CreateFileResponse:
		var resp CreateFileResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	case DirectoryStatResponse:
		var resp DirectoryStatResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	case UploadResponse:
		var resp UploadResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	case ErrorResponse:
		var resp ErrorResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	case UploadDoneResponse:
		var resp UploadDoneResp
		err = c.codex.Decode(pb.Bytes, &resp)
		if err != nil {
			item = &resp
		}

	default:
		return nil, ErrBadType
	}

	if err != nil {
		return nil, err
	}

	return item, nil
}

// Encode will encode a protocol type in to a set of bytes.
func (c *Codex) Encode(what interface{}, version uint8) (*bytes.Buffer, error) {
	msgType, err := c.messageType(what)
	if err != nil {
		return nil, err
	}

	return c.codex.Encode(msgType, version, what)
}

// messageType returns the uint8 type for protocol entry passed to us. It handles
// pointer and non pointer types.
func (c *Codex) messageType(what interface{}) (uint8, error) {
	switch what.(type) {
	case LoginReq, *LoginReq:
		return LoginRequest, nil
	case LogoutReq, *LogoutReq:
		return LogoutRequest, nil
	case CreateProjectReq, *CreateProjectReq:
		return CreateProjectRequest, nil
	case CreateDirectoryReq, *CreateDirectoryReq:
		return CreateDirectoryRequest, nil
	case CreateFileReq, *CreateFileReq:
		return CreateFileRequest, nil
	case DirectoryStatReq, *DirectoryStatReq:
		return DirectoryStatRequest, nil
	case UploadBytesReq, *UploadBytesReq:
		return UploadBytesRequest, nil
	case DoneReq, *DoneReq:
		return DoneRequest, nil
	case LoginResp, *LoginResp:
		return LoginResponse, nil
	case CreateProjectResp, *CreateProjectResp:
		return CreateProjectResponse, nil
	case CreateDirectoryResp, *CreateDirectoryResp:
		return CreateDirectoryResponse, nil
	case CreateFileResp, *CreateFileResp:
		return CreateFileResponse, nil
	case DirectoryStatResp, *DirectoryStatResp:
		return DirectoryStatResponse, nil
	case ErrorResp, *ErrorResp:
		return ErrorResponse, nil
	case UploadDoneResp, *UploadDoneResp:
		return UploadDoneResponse, nil
	default:
		return 0, ErrBadType
	}
}
