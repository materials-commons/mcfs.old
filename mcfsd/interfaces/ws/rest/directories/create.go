package directories

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/app"
	"github.com/materials-commons/mcfs/protocol"
)

func (r *directoriesResource) createDirectory(request *restful.Request, response *restful.Response, user schema.User) (error, interface{}) {
	var req protocol.CreateDirectoryReq

	if err := request.ReadEntity(&req); err != nil {
		return err, nil
	}

	p := app.Directory{
		Name:      req.Name,
		Owner:     user.Name,
		ProjectID: req.ProjectID,
	}

	dir, err := r.dirs.Create(p)
	if err != nil && err != mcerr.ErrExists {
		return err, nil
	}

	return nil, dir
}
