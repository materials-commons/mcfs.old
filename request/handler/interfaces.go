package handler

import (
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
)

type CreateProjectHandler interface {
	Validate(*protocol.CreateProjectReq) bool
	GetProject(name, user string) (*schema.Project, error)
	CreateProject(name, user string) (*schema.Project, error)
}
