package mcfs

import (
	"fmt"
	"github.com/materials-commons/gohandy/marshaling"
	"net"
)

const readBufSize = 1024 * 1024 * 20

type Client struct {
	marshaling.MarshalUnmarshaler
	conn net.Conn
}

type Project struct {
	ProjectID string
	DataDirID string
}

var ErrBadResponseType = fmt.Errorf("Unexpected Response Type")

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
