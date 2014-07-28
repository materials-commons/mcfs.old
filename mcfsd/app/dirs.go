package app

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
)

// Directory represents a directory in the system
type Directory struct {
	Name      string
	ProjectID string
	Owner     string
}

// validDirPath verifies that the directory path starts with the project name.
// It handles both Linux (/) and Windows (\) style slashes.
func (d Directory) validPath(projName string) error {
	slash := strings.Index(d.Name, "/")
	if slash == -1 {
		slash = strings.Index(d.Name, "\\")
	}
	switch {
	case slash == -1:
		return mcerr.ErrInvalid
	case projName != d.Name[:slash]:
		return mcerr.ErrInvalid
	default:
		return nil
	}
}

// DirectoriesService represents the application operations on a directory.
type DirectoriesService interface {
	Create(dir Directory) (*schema.Directory, error)
}

// directoriesService is an implementation of a DirectoriesService.
type directoriesService struct {
	dirs     dai.Dirs
	projects dai.Projects
}

// NewDirectoriesService creates a new directoriesService.
func NewDirectoriesService(dirs dai.Dirs, projects dai.Projects) *directoriesService {
	return &directoriesService{
		dirs:     dirs,
		projects: projects,
	}
}

// Create will create a new directory in the repo or return an existing directory. If a
// new directory is created it returns the directory entry and sets error to nil. If
// an existing directory is returned, then error is set to mcerr.ErrExists. Any other
// error means that no entry was found or created.
func (s *directoriesService) Create(dir Directory) (*schema.Directory, error) {
	proj, err := s.validate(dir)
	if err != nil {
		return nil, err
	}

	return s.createDir(dir, proj)
}

func (s *directoriesService) validate(dir Directory) (*schema.Project, error) {
	proj, err := s.projects.ByID(dir.ProjectID)
	switch {
	case err != nil:
		return nil, err
	case proj.Owner != dir.Owner:
		return nil, mcerr.ErrNoAccess
	case !dir.validPath(proj.Name):
		return nil, mcerr.ErrInvalid
	default:
		return proj, nil
	}
}

// createDir creates a new directory entry if it doesn't exist and the user has permission.
// Otherwise it returns an error.
func (s *directoriesService) createDir(dir Directory) (*schema.Directory, error) {
	// The project exists and the user has permission.
	dataDir, err := s.dirs.ByPath(dir.Name, dir.ProjectID)
	switch {
	case err == mcerr.ErrNotFound:
		// There isn't a matching directory so attempt to create a new one.
		newDir, err := s.createNewDir(dir, proj)
		if err != nil {
			return nil, mcerr.Errorm(mcerr.ErrInvalid, err)
		}
		return newDir, nil
	case err != nil:
		// Lookup failed with an error other than not found.
		return nil, mcerr.Errorm(mcerr.ErrInternal, err)
	default:
		// No error, and the directory already exists, just return it.
		return dataDir, mcerr.ErrExists
	}
}

// createNewDir takes care of creating the directory and attaching it up to
// all the other components and dependencies.
func (s *directoriesService) createNewDir(dir Directory, proj *schema.Project) (*schema.Directory, error) {
	// Each directory has a pointer to its parent directory. Retrieve
	// the parent for the new directory we are creating.
	parent, err := s.getParent(dir)
	if err != nil {
		return nil, err
	}

	datadir := schema.NewDirectory(dir.Name, dir.Owner, dir.ProjectID, parent.ID)
	ddir, err := s.dirs.Insert(&datadir)
	if err != nil {
		return ddir, err
	}

	// Add the directory to the project.
	if err := s.projects.AddDirectories(proj, ddir.ID); err != nil {
		return ddir, err
	}

	return ddir, nil
}

// getParent retrieves the parent directory for a directory path. It does
// this by getting the parent in the path name and then querying the database
// by name for this particular entry. The query is filtered by the project
// which prevents any collisions since a project is a rooted tree.
func (s *directoriesService) getParent(dir Directory) (*schema.Directory, error) {
	var (
		parent *schema.Directory
		err    error
	)
	parentPath := filepath.Dir(dir.Name)
	if parent, err = s.dirs.ByPath(parentPath, dir.ProjectID); err != nil {
		return nil, mcerr.ErrNotFound
	}
	return parent, nil
}
