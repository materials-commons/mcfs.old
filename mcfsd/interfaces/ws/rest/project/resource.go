package project

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/interfaces/db/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
	"github.com/materials-commons/mcfs/protocol"
)

type projectResource struct {
	projects dai.Projects
}

func NewResource(projects dai.Projects) rest.Service {
	return &projectResource{projects: projects}
}

func (r *projectResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/project").Produces(restful.MIME_JSON)
	ws.Route(ws.PUT("/create").To(r.createProject).
		Doc("Creates a new project or retrieves an existing one").
		Reads(protocol.CreateProjectReq{}).
		Writes(protocol.CreateProjectResp{}))
	return ws
}

func (r *projectResource) createProject(request *restful.Request, response *restful.Response) {
	var (
		createProjectReq protocol.CreateProjectReq
		resp             protocol.CreateProjectResp
	)

	if err := request.ReadEntity(&createProjectReq); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, err.Error())
		return
	}

	if !r.validateRequest(&createProjectReq) {
		response.WriteErrorString(http.StatusNotAcceptable, fmt.Sprintf("Invalid project name %s", createProjectReq.Name))
		return
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
			response.WriteErrorString(http.StatusInternalServerError, fmt.Sprintf("Unable to create project: %s", err))
			return
		}
	}

	resp.ProjectID = proj.ID
	resp.DirectoryID = proj.DataDir
	response.WriteEntity(resp)
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
