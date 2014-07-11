package create

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/protocol"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/server/inuse"
	"github.com/materials-commons/mcfs/server/service"
)

// createProjectHandler handles create project request process.
type createProjectHandler struct {
	service *service.Service
}

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
//
// TODO: This method needs to be updated to work with collaboration. Right now a
// user cannot upload files to a project they have access to. Only the owner can
// upload files.
func (r createResource) createProject(request *restful.Request, response *restful.Response) {
	var (
		req  protocol.CreateProjectReq
		proj *schema.Project
		err  error
		resp protocol.CreateProjectResp
		user string
	)
	if err := request.ReadEntity(&req); err != nil {
		response.WriteErrorString(http.StatusNotAcceptable, err.Error())
		return
	}

	s := service.New(service.RethinkDB)
	cph := newCreateProjectHandler(s)

	if !cph.validateRequest(&req) {
		response.WriteErrorString(http.StatusNotAcceptable, fmt.Sprintf("Invalid project name %s", req.Name))
		return
	}

	user = ""

	proj, err = s.Project.ByName(req.Name, user)
	switch {
	case err == nil:
		// Found project
		err = mcerr.ErrExists

	default:
		// Project doesn't exist: Attempt to create a new one.
		proj, err = cph.createNewProject(req.Name, user)
		if err != nil {
			// write an error here
			return
			//return nil, err
		}
	}

	resp.ProjectID = proj.ID
	resp.DirectoryID = proj.DataDir

	// Lock the project so no one else can upload to it.
	if !inuse.Mark(resp.ProjectID) {
		// Project already in use
		// write an error here
		return
		//return nil, mcerr.Errorf(mcerr.ErrInUse, "Project %s is currently in use by someone else.", resp.ProjectID)
	}

	// Save project id so state machine can unlock it at termination.
	//h.projectID = resp.ProjectID
}

func newCreateProjectHandler(service *service.Service) *createProjectHandler {
	return &createProjectHandler{
		service: service,
	}
}

// validateRequest will validate the CreateProjectReq. At the moment this is a very
// simple check to make sure the name is not a file path identifier (ie, contains a '/')
func (cph *createProjectHandler) validateRequest(req *protocol.CreateProjectReq) bool {
	i := strings.Index(req.Name, "/")
	return i == -1
}

// createNewProject creates a new project for the given user.
func (cph *createProjectHandler) createNewProject(name, user string) (*schema.Project, error) {
	project := schema.NewProject(name, "", user)
	newProject, err := cph.service.Project.Insert(&project)
	if err != nil {
		return nil, err
	}
	return newProject, nil
}
