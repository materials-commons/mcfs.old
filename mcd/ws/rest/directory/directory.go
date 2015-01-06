package directory

import (
	"github.com/emicklei/go-restful"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcfs/base/log"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/mcd/dai"
	"github.com/materials-commons/mcfs/mcd/ws/rest"
)

type directoryResource struct {
	log  log15.Logger
	dirs dai.Dirs
}

type directoryRequest struct {
	parentID  string `json:"parentID"`
	name      string `json:"name"`
	projectID string `json:"projectID"`
}

func NewResource() rest.Service {
	return &directoryResource{
		log: log.New("resource", "directory"),
	}
}

func (r *directoryResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/directory").Produces(restful.MIME_JSON)
	ws.Route(ws.POST("").To(rest.RouteHandler(r.createDirectory)).
		Reads(directoryRequest{}).
		Writes(schema.Directory{}))

	return ws
}

func (r *directoryResource) createDirectory(request *restful.Request, response *restful.Response, user schema.User) (error, interface{}) {
	var req directoryRequest
	if err := request.ReadEntity(&req); err != nil {
		return err, nil
	}

	dir := schema.NewDirectory(req.name, user.ID, req.projectID, req.parentID)
	newDir, err := r.dirs.Insert(&dir)
	if err != nil {
		return err, nil
	}

	return nil, &newDir
}
