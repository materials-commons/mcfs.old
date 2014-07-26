package projects

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
	"github.com/materials-commons/mcfs/protocol"
)

type projectResource struct {
	projects dai.Projects
}

func NewResource(projects dai.Projects) rest.Service {
	return &projectResource{projects: projects}
}

func (r *projectResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/project").Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/create").To(rest.RouteHandler(r.createProject)).
		Doc("Creates a new project or retrieves an existing one").
		Reads(protocol.CreateProjectReq{}).
		Writes(protocol.CreateProjectResp{}))
	return ws
}
