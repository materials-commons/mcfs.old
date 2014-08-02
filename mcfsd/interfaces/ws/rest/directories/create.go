package directories

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/app"
	"github.com/materials-commons/mcfs/protocol"
)

func (r *directoriesResource) createDirectory(request *restful.Request, response *restful.Response, user schema.User) error {
	var req protocol.CreateDirectoryReq

	if err := request.ReadEntity(&req); err != nil {
		return err
	}

	p := app.Directory{
		Name:      req.Name,
		Owner:     user.Name,
		ProjectID: req.ProjectID,
	}

	dir, err := r.dirs.Create(p)
	if err != nil && err != mcerr.ErrExists {
		return err
	}

	response.WriteEntity(dir)
	return nil
}
