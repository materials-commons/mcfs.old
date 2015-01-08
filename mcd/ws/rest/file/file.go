package file

import (
	"github.com/emicklei/go-restful"
	"github.com/inconshreveable/log15"
	"github.com/materials-commons/mcfs/base/log"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/mcd/dai"
	"github.com/materials-commons/mcfs/mcd/ws/rest"
)

type fileResource struct {
	log   log15.Logger
	files dai.Files
}

type fileRequest struct {
	parentID  string `json:"parentID"`
	name      string `json:"name"`
	projectID string `json:"projectID"`
}

func NewResource() rest.Service {
	return &fileResource{
		log: log.New("resource", "file"),
	}
}

func (r *fileResource) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/files").Produces(restful.MIME_JSON)
	ws.Route(ws.POST("").To(rest.RouteHandler(r.createFile)).
		Reads(fileRequest{}).
		Writes(schema.File{}))
	return ws
}

func (r *fileResource) createFile(request *restful.Request, response *restful.Response, user schema.User) (error, interface{}) {
	var req fileRequest
	if err := request.ReadEntity(&req); err != nil {
		return err, nil
	}

	file := schema.NewFile("", "")
	newFile, err := r.files.Insert(&file)
	if err != nil {
		return err, nil
	}

	return nil, newFile
}
