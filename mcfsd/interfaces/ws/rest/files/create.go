package files

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/app"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
	"github.com/materials-commons/mcfs/protocol"
)

func (r *filesResource) createFile(request *restful.Request, response *restful.Response) *rest.HTTPError {
	var req protocol.CreateFileReq

	if err := request.ReadEntity(&req); err != nil {
		return rest.HTTPErrore(http.StatusNotAcceptable, err)
	}

	user := request.Attribute("user").(rest.User)
	f := app.File{
		Name:        req.Name,
		ProjectID:   req.ProjectID,
		DirectoryID: req.DirectoryID,
		Checksum:    req.Checksum,
		Size:        req.Size,
		Owner:       user.Name,
	}

	file, err := r.files.Create(f)
	if err != nil {
		return rest.HTTPErrore(http.StatusInternalServerError, err)
	}

	response.WriteEntity(file)
	return nil
}
