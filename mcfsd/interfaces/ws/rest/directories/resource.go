package directories

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
	"github.com/materials-commons/mcfs/protocol"
)

type directoriesResource struct {
	dirs app.DirsService
}

func NewResource(dirs app.DirsService) rest.Service {
	return &directoriesResource{dirs: dirs}
}

func (r *directoriesResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/directories").Produces(restful.MIME_JSON)
	ws.Route(ws.POST("/create").To(rest.RouteHandler(r.createDirectory)).
		Doc("Creates a new directory or retrieves an existing one").
		Reads(protocol.CreateDirectoryReq{}).
		Writes(protocol.CreateDirectoryResp{}))
	return ws
}
