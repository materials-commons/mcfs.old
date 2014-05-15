package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/service"
	"strings"
)

// createProjectHandler handles create project request process.
type createProjectHandler struct {
	dirs     service.Dirs
	projects service.Projects
}

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
//
// TODO: This method needs to be updated to work with collaboration. Right now a
// user cannot upload files to a project they have access to. Only the owner can
// upload files.
func (h *ReqHandler) createProject(req *protocol.CreateProjectReq) (resp *protocol.CreateProjectResp, err error) {
	cph := newCreateProjectHandler()

	if !cph.validateRequest(req) {
		return nil, mc.Errorf(mc.ErrInvalid, "Invalid project name %s", req.Name)
	}

	proj, err := cph.projects.ByName(req.Name, h.user)
	switch {
	case err == nil:
		// Found project
		resp := &protocol.CreateProjectResp{
			ProjectID: proj.ID,
			DataDirID: proj.DataDir,
		}
		return resp, mc.ErrExists

	default:
		// Project doesn't exist. Create a new one and return it.
		p, err := cph.createNewProject(req.Name, h.user)
		if err != nil {
			return nil, err
		}
		resp := &protocol.CreateProjectResp{
			ProjectID: p.ID,
			DataDirID: p.DataDir,
		}
		return resp, nil
	}
}

func newCreateProjectHandler() *createProjectHandler {
	return &createProjectHandler{
		dirs:     service.NewDirs(service.RethinkDB),
		projects: service.NewProjects(service.RethinkDB),
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
	newProject, err := cph.projects.Insert(&project)
	if err != nil {
		return nil, err
	}
	return newProject, nil
}
