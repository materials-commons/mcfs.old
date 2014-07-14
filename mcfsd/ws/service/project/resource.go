package project

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/ws/service"
)

type projectResource struct{}

func NewResource() service.REST {
	return &projectResource{}
}

func (r *projectResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/project").Produces(restful.MIME_JSON)
	return ws
}
