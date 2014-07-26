package app

import (
	"strings"

	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
)

// Project represents a materials commons project.
type Project struct {
	Name  string
	Owner string
}

// isValid sanity checks a project object.
func (p Project) isValid() error {
	switch {
	case p.Name == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No name")
	case p.Owner == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No owner")
	case strings.Index(p.Name, "/") != -1:
		return mcerr.Errorf(mcerr.ErrInvalid, "Invalid name")
	default:
		return nil
	}
}

// ProjectsService represents the application operations on a project.
type ProjectsService interface {
	Create(project Project) (*schema.Project, error)
}

// projectsService is an implementation of ProjectsService
type projectsService struct {
	projects dai.Projects
}

// NewProjectsService returns a new projectsService.
func NewProjectsService(projects dai.Projects) *projectsService {
	return &projectsService{
		projects: projects,
	}
}

// Create will create a new project in the repo or return an existing project.
// If a new project created it returns the project entry and sets error to
// nil. If an existing project is returned, then error is set to
// mcerr.ErrExists. Any other error means that no project was found or created.
func (p *projectsService) Create(project Project) (*schema.Project, error) {
	if err := project.isValid(); err != nil {
		return nil, err
	}

	return p.createProject(project)
}

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
//
// TODO: This method needs to be updated to work with collaboration. Right now a
// non-owner user cannot upload files to a project they have access to. Only the
// owner of the project can upload files.
func (p *projectsService) createProject(project Project) (*schema.Project, error) {
	proj, err := p.projects.ByName(project.Name, project.Name)
	if err != nil {
		// Project doesn't exist: Attempt to create a new one.
		return p.createNewProject(project.Name, project.Name)
	}

	// No error means we found the project
	return proj, mcerr.ErrExists
}

// createNewProject creates a new project for the given user.
func (p *projectsService) createNewProject(name, user string) (*schema.Project, error) {
	project := schema.NewProject(name, "", user)
	newProject, err := p.projects.Insert(&project)
	if err != nil {
		return nil, err
	}
	return newProject, nil
}
