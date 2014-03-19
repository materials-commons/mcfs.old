package handler

import (
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
)

// CreateProjectHandler defines the interface for creating project.
type CreateProjectHandler interface {
	Validate(req *protocol.CreateProjectReq) bool
	GetProject(name, user string) (*schema.Project, error)
	CreateProject(name, user string) (*schema.Project, error)
}

// CreateDirHandler defines the interface for creating a directory.
type CreateDirHandler interface {
	GetProject(id string) (*schema.Project, error)
	GetDataDir(req *protocol.CreateDirReq) (*schema.Directory, error)
	GetParent(path string) (*schema.Directory, error)
	CreateDir(req *protocol.CreateDirReq, user, parentID string) (*schema.Directory, error)
}

// CreateFileHandler defines the interface for creating a file.
type CreateFileHandler interface {
	Validate(req *protocol.CreateFileReq) error
	CreateFile(req *protocol.CreateFileReq, user string) (*schema.Directory, error)
}
