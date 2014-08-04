package projects

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/protocol"
)

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
func (r *projectResource) createProject(request *restful.Request, response *restful.Response, user schema.User) (error, interface{}) {
	var req protocol.CreateProjectReq

	if err := request.ReadEntity(&req); err != nil {
		return err, nil
	}

	project, err := r.projects.Create(req.Name, user.Name)
	if err != nil && err != mcerr.ErrExists {
		return err, nil
	}

	return nil, project
}
