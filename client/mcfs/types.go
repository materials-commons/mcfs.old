package mcfs

import (
	"fmt"
	"github.com/materials-commons/gohandy/marshaling"
	"net"
)

const readBufSize = 1024 * 1024 * 20

// Client represents a client connection to the sever.
type Client struct {
	marshaling.MarshalUnmarshaler
	conn net.Conn
}

// Project holds ids the server uses for a project.
type Project struct {
	ProjectID string
	DataDirID string
}

// ErrBadResponseType is an error where the server sent us a response
// we do not recognize.
var ErrBadResponseType = fmt.Errorf("unexpected response type")

// DataFileUpload tracks a particular upload request.
type DataFileUpload struct {
	ProjectID     string
	DataDirID     string
	DataFileID    string
	Path          string
	Size          int64
	Checksum      string
	BytesUploaded int64
	Err           error
}
