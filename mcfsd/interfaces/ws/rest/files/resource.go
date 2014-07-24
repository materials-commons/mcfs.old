package files

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/app"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
	"github.com/materials-commons/mcfs/protocol"
)

type filesResource struct {
	files app.FilesService
}

// NewResource returns a new Resource.
func NewResource(files app.FilesService) *filesResource {
	return &filesResource{
		files: files,
	}
}

//
func (r *filesResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/files").Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/create").To(rest.RouteHandler(r.createFile)).
		Doc("Creates a new file or retrieves an existing one").
		Reads(protocol.CreateFileReq{}).
		Writes(protocol.CreateFileResp{}))
	return ws
}
