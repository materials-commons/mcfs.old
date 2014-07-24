package project

import (
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
	"github.com/materials-commons/mcfs/protocol"
)

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
//
// TODO: This method needs to be updated to work with collaboration. Right now a
// non-owner user cannot upload files to a project they have access to. Only the
// owner of the project can upload files.
func (r *projectResource) createProject(request *restful.Request, response *restful.Response) *rest.HTTPError {
	var (
		createProjectReq protocol.CreateProjectReq
		resp             protocol.CreateProjectResp
	)

	if err := request.ReadEntity(&createProjectReq); err != nil {
		return rest.HTTPErrorm(http.StatusNotAcceptable, err.Error())
	}

	if !r.validateRequest(&createProjectReq) {
		return rest.HTTPErrorf(http.StatusNotAcceptable, "Invalid project name %s", createProjectReq.Name)
	}

	user := request.Attribute("user").(rest.User)

	proj, err := r.projects.ByName(createProjectReq.Name, user.Name)
	switch {
	case err == nil:
		// Found project
		err = mcerr.ErrExists

	default:
		// Project doesn't exist: Attempt to create a new one.
		proj, err = r.createNewProject(createProjectReq.Name, user.Name)
		if err != nil {
			return rest.HTTPErrorf(http.StatusInternalServerError, "Unable to create project: %s", err)
		}
	}

	// TODO: Should we just return the schema.Project?
	resp.ProjectID = proj.ID
	resp.DirectoryID = proj.DataDir
	response.WriteEntity(resp)
	return nil
}

// validateRequest will validate the CreateProjectReq. At the moment this is a very
// simple check to make sure the name is not a file path identifier (ie, contains a '/')
func (r *projectResource) validateRequest(req *protocol.CreateProjectReq) bool {
	i := strings.Index(req.Name, "/")
	return i == -1
}

// createNewProject creates a new project for the given user.
func (r *projectResource) createNewProject(name, user string) (*schema.Project, error) {
	project := schema.NewProject(name, "", user)
	newProject, err := r.projects.Insert(&project)
	if err != nil {
		return nil, err
	}
	return newProject, nil
}
