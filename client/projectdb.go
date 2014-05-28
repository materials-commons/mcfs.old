package materials

import (
	"encoding/json"
	"fmt"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/client/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// ProjectDB contains a list of user projects and information that
// is needed by the methods to load the projects file.
type ProjectDB struct {
	path     string
	projects []*Project
	mutex    sync.RWMutex
}

var currentUserDB *ProjectDB
var currentUserDBOnce sync.Once

// CurrentUserProjectDB opens the project database for a user contained in
// $HOME/.materials/projectdb. Only opens the database once per process,
// regardless of how many times it is called.
func CurrentUserProjectDB() *ProjectDB {
	currentUserDBOnce.Do(func() {
		projectsPath := filepath.Join(config.Config.User.DotMaterialsPath(), "projectdb")
		var err error
		currentUserDB, err = OpenProjectDB(projectsPath)
		if err != nil {
			panic(fmt.Sprintf("Unable to open current users projectsdb: %s", projectsPath))
		}
	})
	return currentUserDB
}

// OpenProjectDB loads projects from the database directory at path. Project files are
// JSON files ending with a .project extension.
func OpenProjectDB(path string) (*ProjectDB, error) {
	projectDB := ProjectDB{path: path}
	err := projectDB.loadProjects()
	if err != nil {
		return nil, err
	}
	return &projectDB, err
}

// Reload re-reads and loads the projects file.
func (p *ProjectDB) Reload() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.loadProjects()
}

// loadProjects reads the projects directory, and loads each *.project file found in it.
func (p *ProjectDB) loadProjects() error {
	if !file.IsDir(p.path) {
		return fmt.Errorf("projectdb must be a directory: '%s'", p.path)
	}

	finfos, err := ioutil.ReadDir(p.path)
	if err != nil {
		return err
	}

	p.projects = []*Project{}
	for _, finfo := range finfos {
		if isProjectFile(finfo) {
			proj, err := readProjectFile(filepath.Join(p.path, finfo.Name()))
			if err == nil {
				proj.OpenDB()
				p.projects = append(p.projects, proj)
			}
		}
	}

	return nil
}

// isProjectFile tests if a FileInfo project points to a project file.
func isProjectFile(finfo os.FileInfo) bool {
	if !finfo.IsDir() {
		if ext := filepath.Ext(finfo.Name()); ext == ".project" {
			return true
		}
	}

	return false
}

// readProjectFile reads a a project file, parses the JSON in a Project.
func readProjectFile(filepath string) (*Project, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(b, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// Projects returns the list of loaded projects.
func (p *ProjectDB) Projects() []*Project {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.projects
}

/*********************************************************************
 * Add, Remove and Update should only call low level routines, never
 * public routes since the public routines will take out a lock and
 * you will end up in a deadlock situation.
 *********************************************************************/

// Add adds a new project to and writes the corresponding project file.
func (p *ProjectDB) Add(proj Project) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, index := p.find(proj.Name); index != -1 {
		return fmt.Errorf("project already exists: %s", proj.Name)
	}

	if err := p.writeProject(proj); err != nil {
		return err
	}

	p.projects = append(p.projects, &proj)
	return nil
}

// Remove removes a project and its file.
func (p *ProjectDB) Remove(projectName string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	projects, projectFound := p.projectsExceptFor(projectName)

	// We found the entry to remove, so we attempt to remove the project file.
	if projectFound {
		if err := os.Remove(p.projectFilePath(projectName)); err != nil {
			return err
		}
	}

	p.projects = projects
	return nil
}

// Update updates an existing project and its file.
func (p *ProjectDB) Update(f func() *Project) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	proj := f()
	if _, index := p.find(proj.Name); index != -1 {
		if err := p.writeProject(*proj); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("project not found: %s", proj.Name)
}

// projectsExceptFor returns a new list of projects except for the project
// matching projectName. It returns true if it found a project matching
// projectName.
func (p *ProjectDB) projectsExceptFor(projectName string) ([]*Project, bool) {
	projects := []*Project{}
	found := false
	for _, project := range p.projects {
		if project.Name != projectName {
			projects = append(projects, project)
		} else {
			found = true
		}
	}
	return projects, found
}

// writeProject writes a project to a project file.
func (p *ProjectDB) writeProject(project Project) error {
	b, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return err
	}

	filename := p.projectFilePath(project.Name)
	return ioutil.WriteFile(filename, b, os.ModePerm)
}

// projectFilePath creates the path to a projects file.
func (p *ProjectDB) projectFilePath(projectName string) string {
	return filepath.Join(p.path, projectName+".project")
}

// Exists returns true if there is a project matching
// the given Name.
func (p *ProjectDB) Exists(projectName string) bool {
	_, found := p.Find(projectName)
	return found
}

// Find returns (Project, true) if the project is found otherwise
// it returns (Project{}, false).
func (p *ProjectDB) Find(projectName string) (*Project, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	project, index := p.find(projectName)
	return project, index != -1
}

// find returns (Project, index) where index is -1 if
// the project wasn't found, otherwise it is the index
// in the Projects array.
func (p *ProjectDB) find(projectName string) (*Project, int) {
	// Never put a lock in this routine.
	for index, project := range p.projects {
		if project.Name == projectName {
			return project, index
		}
	}

	return nil, -1
}
