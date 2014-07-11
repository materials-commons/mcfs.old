package create

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/base/protocol"
)

type createResource struct {
}

func NewResource(container *restful.Container) error {
	createResource := createResource{}
	createResource.register(container)
	return nil
}

func (r createResource) register(container *restful.Container) {
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

	container.Add(ws)
}

func (r createResource) createFile(request *restful.Request, response *restful.Response) {

}

func (r createResource) createDirectory(request *restful.Request, response *restful.Response) {

}
