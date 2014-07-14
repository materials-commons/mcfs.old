package create

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/ws/service"
	"github.com/materials-commons/mcfs/protocol"
)

type createResource struct {
}

func NewResource() service.REST {
	return &createResource{}
}

// WebService creates an instance of the create webservice.
func (r *createResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/create").Produces(restful.MIME_JSON)
	ws.Route(ws.PUT("/file").To(r.createFile).
		Doc("Creates a new file or retrieves an existing one").
		Reads(protocol.CreateFileReq{}).
		Writes(protocol.CreateFileResp{}))
	ws.Route(ws.PUT("/directory").To(r.createDirectory).
		Doc("Creates a new directory or retrieves an existing one").
		Reads(protocol.CreateDirectoryReq{}).
		Writes(protocol.CreateDirectoryResp{}))
	ws.Route(ws.PUT("/project").To(r.createProject).
		Doc("Creates a new project or retrieves an existing one").
		Reads(protocol.CreateProjectReq{}).
		Writes(protocol.CreateProjectResp{}))

	return ws
}

func (r *createResource) createFile(request *restful.Request, response *restful.Response) {

}

func (r *createResource) createDirectory(request *restful.Request, response *restful.Response) {

}
