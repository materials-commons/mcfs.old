package app

import (
	"strings"
	"time"

	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/mcd/dai"
)

// Note is a user note entry.
type Note struct {
	ID        string    `json:"id"`
	Birthtime time.Time `json:"birthtime"`
	MTime     time.Time `json:"mtime"`
	Date      int       `json:"date"`
	Message   string    `json:"message"`
	Who       string    `json:"who"`
}

// Tag is a user tag.
type Tag struct {
	Name string `json:"name"`
}

// Draft is a draft provenance entry.
type Draft struct {
	ID string `json:"id"`
}

// Review is a request to review an item.
type Review struct {
	ID          string    `json:"id"`
	Birthtime   time.Time `json:"birthtime"`
	ItemID      string    `json:"item_id"`
	ItemName    string    `json:"item_name"`
	ItemType    string    `json:"item_type"`
	ProjectID   string    `json:"project_id"`
	RequestedBy string    `json:"requested_by"`
	RequestTo   string    `json:"request_to"`
	Status      string    `json:"status"`
	Notes       []Note    `json:"notes"`
}

// ACL controls access to a dataset.
type ACL struct {
	Dataset     string `json:"dataset"`
	Permissions string `json:"permissions"`
}

// Access contains the access permissions for a user.
type Access struct {
	User string `json:"user"`
	ACLs []ACL  `json:"acls"`
}

// Project is a user project in the system. A project holds
// the files, directories and meta data. A project controls
// access and visibility.
type Project struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	Birthtime   time.Time `json:"birthtime"`
	MTime       time.Time `json:"mtime"`
	Notes       []Note    `json:"notes"`
	Tags        []Tag     `json:"tags"`
	Reviews     []Review  `json:"reviews"`
	MyTags      []Tag     `json:"mytags"`
	Drafts      []Draft   `json:"drafts"`
	Access      []Access  `json:"access"`
}

// ProjectsService represents the application operations on a project.
type ProjectsService interface {
	Create(name, owner string) (*schema.Project, error)
	Get(id string) (Project, error)
	ForUser(user string) ([]Project, error)
	Update(project Project) error
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

// validArgs sanity checks the project name and owner arguments.
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

// Get retrieves a single project.
func (p *projectsService) Get(id string) (Project, error) {
	proj, err := p.projects.ByID(id)
	if err != nil {
		return Project{}, err
	}
	project := Project{
		ID:          proj.ID,
		Name:        proj.Name,
		Description: proj.Description,
		Owner:       proj.Owner,
		Birthtime:   proj.Birthtime,
		MTime:       proj.MTime,
	}
	return project, nil
}

// ForUser retrieves all the projects that a user has access to.
func (p *projectsService) ForUser(user string) ([]Project, error) {
	var projects []schema.Project
	projects, err := p.projects.ForUser(user)
	if err != nil {
		return nil, err
	}

	projs := p.convertProjects(projects)
	return projs, nil
}

// makeProject takes a schema.Project and turns into a Project.
func (p *projectsService) convertProjects(projects []schema.Project) []Project {
	var projs []Project
	for _, proj := range projects {
		project := Project{
			ID:          proj.ID,
			Name:        proj.Name,
			Description: proj.Description,
			Owner:       proj.Owner,
			Birthtime:   proj.Birthtime,
			MTime:       proj.MTime,
		}
		projs = append(projs, project)
	}
	return projs
}
