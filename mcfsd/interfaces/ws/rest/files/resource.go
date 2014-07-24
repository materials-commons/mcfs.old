package files

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
	"github.com/materials-commons/mcfs/protocol"
)

type filesResource struct {
	files    dai.Files
	dirs     dai.Dirs
	projects dai.Projects
	groups   dai.Groups
}

// NewResource returns a new Resource.
func NewResource(files dai.Files, dirs dai.Dirs, projects dai.Projects, groups dai.Groups) *filesResource {
	return &filesResource{
		files:    files,
		dirs:     dirs,
		projects: projects,
		groups:   groups,
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
