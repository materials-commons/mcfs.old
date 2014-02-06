package handler

import (
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
)

type CreateProjectHandler interface {
	Validate(req *protocol.CreateProjectReq) bool
	GetProject(name, user string) (*schema.Project, error)
	CreateProject(name, user string) (*schema.Project, error)
}

type CreateDirHandler interface {
	GetProject(id string) (*schema.Project, error)
	GetDataDir(req *protocol.CreateDirReq) (*schema.DataDir, error)
	GetParent(path string) (*schema.DataDir, error)
	CreateDir(req *protocol.CreateDirReq, user, parentID string) (*schema.DataDir, error)
}

type CreateFileHandler interface {
	Validate(req *protocol.CreateFileReq) error
	CreateFile(req *protocol.CreateFileReq, user string) (*schema.DataDir, error)
}
