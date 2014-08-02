package projects

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"

	"github.com/materials-commons/mcfs/mcfsd/app"
	"github.com/materials-commons/mcfs/protocol"
)

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
//
// TODO: This method needs to be updated to work with collaboration. Right now a
// non-owner user cannot upload files to a project they have access to. Only the
// owner of the project can upload files.
func (r *projectResource) createProject(request *restful.Request, response *restful.Response, user schema.User) (error, interface{}) {
	var req protocol.CreateProjectReq

	if err := request.ReadEntity(&req); err != nil {
		return err, nil
	}

	p := app.Project{
		Name:  req.Name,
		Owner: user.Name,
	}

	project, err := r.projects.Create(p)
	if err != nil && err != mcerr.ErrExists {
		return err, nil
	}

	return nil, project
}
