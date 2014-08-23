package create

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/common/schema" // createProjectHandler handles create project request process.
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/service"
	"github.com/materials-commons/mcfs/mcfsd/ws"
	"github.com/materials-commons/mcfs/protocol"
)

type createProjectHandler struct {
	service *service.Service
}

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
//
// TODO: This method needs to be updated to work with collaboration. Right now a
// user cannot upload files to a project they have access to. Only the owner can
// upload files.
func (r *createResource) createProject(request *restful.Request, response *restful.Response) {
	var (
		req  protocol.CreateProjectReq
		resp protocol.CreateProjectResp
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

	user := request.Attribute("user").(ws.User)

	proj, err := s.Project.ByName(req.Name, user.Name)
	switch {
	case err == nil:
		// Found project
		err = mcerr.ErrExists

	default:
		// Project doesn't exist: Attempt to create a new one.
		proj, err = cph.createNewProject(req.Name, user.Name)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, fmt.Sprintf("Unable to create project: %s", err))
			return
		}
	}

	resp.ProjectID = proj.ID
	resp.DirectoryID = proj.DataDir
	response.WriteEntity(resp)
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
