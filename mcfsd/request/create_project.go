package request

import (
	"strings"

	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/protocol"
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
func (h *ReqHandler) createProject(req *protocol.CreateProjectReq) (*protocol.CreateProjectResp, error) {
	var (
		proj *schema.Project
		resp protocol.CreateProjectResp
		err  error
	)
	cph := newCreateProjectHandler(h.service)

	if !cph.validateRequest(req) {
		return nil, mcerr.Errorf(mcerr.ErrInvalid, "Invalid project name %s", req.Name)
	}

	proj, err = h.service.Project.ByName(req.Name, h.user)
	switch {
	case err == nil:
		// Found project
		err = mcerr.ErrExists

	default:
		// Project doesn't exist: Attempt to create a new one.
		proj, err = cph.createNewProject(req.Name, h.user)
		if err != nil {
			return nil, err
		}
	}

	resp.ProjectID = proj.ID
	resp.DirectoryID = proj.DataDir

	// Save project id so state machine can unlock it at termination.
	h.projectID = resp.ProjectID
	return &resp, err
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
