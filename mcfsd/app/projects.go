package app

import (
	"strings"
	"time"

	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
)

// Project is a user project in the system. A project holds
// the files, directories and meta data. A project controls
// access and visibility.
type Project struct {
	ID          string          `json:"id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Owner       string          `json:"owner"`
	Birthtime   time.Time       `json:"birthtime"`
	MTime       time.Time       `json:"mtime"`
	Notes       []schema.Note   `json:"notes"`
	Tags        []schema.Tag    `json:"tags"`
	Reviews     []schema.Review `json:"reviews"`
	MyTags      []schema.Tag    `json:"mytags"`
	Drafts      []schema.Draft  `json:"drafts"`
	Samples     []schema.Sample `json:"samples"`
	Groups      []schema.Group  `json:"groups"`
}

// ProjectsService represents the application operations on a project.
type ProjectsService interface {
	Create(name, owner string) (*schema.Project, error)
	Get(id string) (*Project, error)
	ForUser(user string) ([]Project, error)
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

// isValid sanity checks a project object.
func (p projectsService) validArgs(name, owner string) error {
	switch {
	case name == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No name")
	case owner == "":
		return mcerr.Errorf(mcerr.ErrInvalid, "No owner")
	case strings.Index(name, "/") != -1:
		return mcerr.Errorf(mcerr.ErrInvalid, "Invalid name")
	default:
		return nil
	}
}

// Create will create a new project in the repo or return an existing project.
// If a new project created it returns the project entry and sets error to
// nil. If an existing project is returned, then error is set to
// mcerr.ErrExists. Any other error means that no project was found or created.
func (p *projectsService) Create(name, owner string) (*schema.Project, error) {
	if err := p.validArgs(name, owner); err != nil {
		return nil, err
	}

	return p.createProject(name, owner)
}

// createProject will create a new project or return an existing project. Only owners
// of a project can create or access an existing project.
//
// TODO: This method needs to be updated to work with collaboration. Right now a
// non-owner user cannot upload files to a project they have access to. Only the
// owner of the project can upload files.
func (p *projectsService) createProject(name, owner string) (*schema.Project, error) {
	proj, err := p.projects.ByName(name, owner)
	if err != nil {
		// Project doesn't exist: Attempt to create a new one.
		return p.createNewProject(name, owner)
	}

	// No error means we found the project
	return proj, mcerr.ErrExists
}

// createNewProject creates a new project for the given owner.
func (p *projectsService) createNewProject(name, owner string) (*schema.Project, error) {
	project := schema.NewProject(name, "", owner)
	newProject, err := p.projects.Insert(&project)
	if err != nil {
		return nil, err
	}
	return newProject, nil
}

func (p *projectsService) Get(id string) (*Project, error) {
	proj, err := p.projects.ByID(id)
	if err != nil {
		return nil, err
	}

	project := &Project{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Owner:       proj.Owner,
		Birthtime:   proj.Birthtime,
		MTime:       proj.MTime,
		Samples:     p.getSamples(id),
		Reviews:     p.getReviews(id),
	}

	return project, nil
}

func (p *projectsService) ForUser(user string) ([]Project, error) {
	return nil, nil
}

func (p *projectsService) getSamples(id string) []schema.Sample {
	samples, err := p.projects.GetSamples(id)
	if err != nil {
		return []schema.Sample{}
	}
	return samples
}

func (p *projectsService) getReviews(id string) []schema.Review {
	reviews, err := p.projects.GetReviews(id)
	if err != nil {
		return []schema.Review{}
	}
	return reviews
}
