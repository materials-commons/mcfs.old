/*
Package protocol contains the protocol definitions for uploads/downloads.
*/
package protocol

import (
	"encoding/gob"
	"github.com/materials-commons/mcfs/base/dir"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/schema"
	"time"
)

func init() {
	gob.Register(Response{})
	gob.Register(Request{})

	gob.Register(UploadReq{})
	gob.Register(UploadResp{})

	gob.Register(DownloadReq{})
	gob.Register(DownloadResp{})

	gob.Register(MoveReq{})
	gob.Register(MoveResp{})

	gob.Register(DeleteReq{})
	gob.Register(DeleteResp{})

	gob.Register(SendReq{})
	gob.Register(SendResp{})

	gob.Register(StatReq{})
	gob.Register(StatResp{})

	gob.Register(EndReq{})
	gob.Register(EndResp{})

	gob.Register(CreateFileReq{})
	gob.Register(CreateDirReq{})
	gob.Register(CreateProjectReq{})

	gob.Register(CreateProjectResp{})
	gob.Register(CreateResp{})

	gob.Register(LoginReq{})
	gob.Register(LoginResp{})
	gob.Register(LogoutReq{})
	gob.Register(LogoutResp{})

	gob.Register(StartResp{})

	gob.Register(CloseReq{})
	gob.Register(IndexReq{})
	gob.Register(DoneReq{})
	gob.Register(DoneResp{})

	gob.Register(LookupReq{})

	gob.Register(schema.File{})
	gob.Register(schema.Directory{})
	gob.Register(schema.Project{})

	gob.Register(StatProjectReq{})
	gob.Register(StatProjectResp{})
	gob.Register(dir.FileInfo{})
}

// Request defines the request being made.
type Request struct {
	Req interface{}
}

// ItemType is the type of item.
type ItemType int

const (
	// DataDir directory
	DataDir ItemType = iota

	// DataFile file
	DataFile

	//Project project
	Project

	// DataSet dataset
	DataSet
)

var itemTypes = map[ItemType]bool{
	DataDir:  true,
	DataFile: true,
	Project:  true,
	DataSet:  true,
}

// ValidItemType checks if the ItemType is known.
func ValidItemType(t ItemType) bool {
	return itemTypes[t]
}

// Response contains the response to a given request.
type Response struct {
	Status        mcerr.ErrorCode
	StatusMessage string
	Resp          interface{}
}

// UploadReq is an upload request.
type UploadReq struct {
	DataFileID string
	Checksum   string
	Size       int64
}

// UploadResp is an upload response.
type UploadResp struct {
	DataFileID string
	Offset     int64
}

// DownloadReq is an download request.
type DownloadReq struct {
	Type ItemType
	ID   string
}

// DownloadResp is an download response.
type DownloadResp struct {
	Ok bool
}

// MoveReq is a file or directory move request.
type MoveReq struct {
}

// MoveResp is a file or directory move response.
type MoveResp struct {
}

// DeleteReq is a file or directory move request.
type DeleteReq struct {
}

// DeleteResp is a file or directory move response.
type DeleteResp struct {
}

// SendReq is request to send a set of bytes from a file.
type SendReq struct {
	DataFileID string
	Bytes      []byte
}

// SendResp is the response to the SenReq.
type SendResp struct {
	BytesWritten int
}

// StatReq is a status request to get information on a datafile.
type StatReq struct {
	DataFileID string
}

// StatResp is the response for a StatReq.
type StatResp struct {
	DataFileID string
	Name       string
	DataDirs   []string
	Checksum   string
	Size       int64
	Birthtime  time.Time
	MTime      time.Time
}

// EndReq specifies that no more requests are coming.
type EndReq struct {
}

// EndResp is the response to a EndReq.
type EndResp struct {
	Ok bool
}

// CreateFileReq requests the creation of a new file on the server.
type CreateFileReq struct {
	ProjectID string
	DataDirID string
	Name      string
	Checksum  string
	Size      int64
}

// CreateDirReq requests the creation of a new directory on the server.
type CreateDirReq struct {
	ProjectID string
	Path      string
}

// CreateProjectReq requests the creation of a new project on the server.
type CreateProjectReq struct {
	Name string
}

// CreateProjectResp is the response to creating a new project on the server. It
// returns the id for the project and its top level directory.
type CreateProjectResp struct {
	ProjectID string
	DataDirID string
}

// CreateResp is a generic response to a create that returns the id of the item created.
type CreateResp struct {
	ID string
}

// LoginReq login request.
type LoginReq struct {
	User   string
	APIKey string
}

// LoginResp login response.
type LoginResp struct{}

// LogoutReq logout request.
type LogoutReq struct{}

// LogoutResp logout response.
type LogoutResp struct{}

// StartResp response to a start.
type StartResp struct {
	Ok bool
}

// CloseReq close request.
type CloseReq struct{}

// IndexReq index request.
type IndexReq struct{}

// DoneReq done request.
type DoneReq struct{}

// DoneResp done response.
type DoneResp struct{}

// LookupReq lookup request.
type LookupReq struct {
	Field     string
	Value     string
	Type      string
	LimitToID string
}

// StatProjectReq project entries request.
type StatProjectReq struct {
	Name string
	ID   string
	Base string
}

// StatProjectResp response to a ProjectEntriesReq.
type StatProjectResp struct {
	ProjectID string
	Entries   []dir.FileInfo
}
