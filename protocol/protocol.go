package protocol

import (
	"encoding/gob"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
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

	gob.Register(schema.DataFile{})
	gob.Register(schema.DataDir{})
	gob.Register(schema.Project{})

	gob.Register(ProjectEntriesReq{})
	gob.Register(ProjectEntriesResp{})
	gob.Register(ProjectEntry{})
}

type Request struct {
	Req interface{}
}

type ItemType int

const (
	DataDir ItemType = iota
	DataFile
	Project
	DataSet
)

var itemTypes = map[ItemType]bool{
	DataDir:  true,
	DataFile: true,
	Project:  true,
	DataSet:  true,
}

func ValidItemType(t ItemType) bool {
	return itemTypes[t]
}

type Response struct {
	Status        mc.ErrorCode
	StatusMessage string
	Resp          interface{}
}

type UploadReq struct {
	DataFileID string
	Checksum   string
	Size       int64
}

type UploadResp struct {
	DataFileID string
	Offset     int64
}

type DownloadReq struct {
	Type ItemType
	ID   string
}

type DownloadResp struct {
	Ok bool
}

type MoveReq struct {
}

type MoveResp struct {
}

type DeleteReq struct {
}

type DeleteResp struct {
}

type SendReq struct {
	DataFileID string
	Bytes      []byte
}

type SendResp struct {
	BytesWritten int
}

type StatReq struct {
	DataFileID string
}

type StatResp struct {
	DataFileID string
	Name       string
	DataDirs   []string
	Checksum   string
	Size       int64
	Birthtime  time.Time
	MTime      time.Time
}

type EndReq struct {
}

type EndResp struct {
	Ok bool
}

type CreateFileReq struct {
	ProjectID string
	DataDirID string
	Name      string
	Checksum  string
	Size      int64
}

type CreateDirReq struct {
	ProjectID string
	Path      string
}

type CreateProjectReq struct {
	Name string
}

type CreateProjectResp struct {
	ProjectID string
	DataDirID string
}

type CreateResp struct {
	ID string
}

type LoginReq struct {
	User   string
	ApiKey string
}

type LoginResp struct{}

type LogoutReq struct{}
type LogoutResp struct{}

type StartResp struct {
	Ok bool
}

type CloseReq struct{}
type IndexReq struct{}
type DoneReq struct{}
type DoneResp struct{}

type LookupReq struct {
	Field     string
	Value     string
	Type      string
	LimitToID string
}

type ProjectEntriesReq struct {
	Name string
}

type ProjectEntry struct {
	DataDirID        string `gorethink:"datadir_id"`
	DataDirName      string `gorethink:"datadir_name"`
	DataFileID       string `gorethink:"id"`
	DataFileName     string `gorethink:"name"`
	DataFileSize     int64  `gorethink:"size"`
	DataFileChecksum string `gorethink:"checksum"`
}

type ProjectEntriesResp struct {
	ProjectID string
	Entries   []ProjectEntry
}
