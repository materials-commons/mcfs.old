package request

import (
	"path/filepath"
	"strings"

	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/server/dai"
)

// createDirHandler is an internal handler for creating a directory.
// It holds state information needed to create a new directory entry.
type createDirHandler struct {
	req     *protocol.CreateDirReq
	user    string
	proj    *schema.Project
	dai *dai.Service
}

// createDir creates a new directory entry if it doesn't exist and the user has permission.
// Otherwise it returns an error.
func (h *ReqHandler) createDir(req *protocol.CreateDirReq) (resp *protocol.CreateResp, err error) {
	cdh := newCreateDirHandler(req, h.user, h.dai)

	// Get the project since a directory is added to a project.
	cdh.proj, err = h.dai.Project.ByID(req.ProjectID)
	switch {
	case err != nil:
		// A bad projectID was passed to us
		return nil, mcerr.Errorf(mcerr.ErrInvalid, "Bad projectID %s", req.ProjectID)
	case cdh.proj.Owner != h.user:
		// A valid project but the user doesn't have permission to add an entry
		// to this project.
		return nil, mcerr.Errorf(mcerr.ErrNoAccess, "Access to project %s not allowed", req.ProjectID)
	case !validDirPath(cdh.proj.Name, req.Path):
		// The format for the path is incorrect.
		return nil, mcerr.Errorf(mcerr.ErrInvalid, "Invalid directory path %s", req.Path)
	default:
		// The project exists and the user has permission.
		dataDir, err := h.dai.Dir.ByPath(req.Path, req.ProjectID)
		switch {
		case err == mcerr.ErrNotFound:
			// There isn't a matching directory so attempt to create a new one.
			dataDir, err := cdh.createNewDir()
			if err != nil {
				return nil, mcerr.Errorm(mcerr.ErrInvalid, err)
			}
			return &protocol.CreateResp{ID: dataDir.ID}, nil
		case err != nil:
			// Lookup failed with an error other than not found.
			return nil, mcerr.Errorm(mcerr.ErrNotFound, err)
		default:
			// No error, and the directory already exists, just return it.
			return &protocol.CreateResp{ID: dataDir.ID}, nil
		}
	}
}

// newCreateDirHandler creates a new instance of an createDirHandler. The constructor
// also sets up the dirs and projects models.
func newCreateDirHandler(req *protocol.CreateDirReq, user string, dai *dai.Service) *createDirHandler {
	return &createDirHandler{
		req:     req,
		user:    user,
		dai: dai,
	}
}

// createNewDir takes care of creating the directory and attaching it up to
// all the other components and dependencies.
func (cdh *createDirHandler) createNewDir() (*schema.Directory, error) {
	// Each directory has a pointer to its parent directory. Retrieve
	// the parent for the new directory we are creating.
	parent, err := cdh.getParent()
	if err != nil {
		return nil, err
	}

	datadir := schema.NewDirectory(cdh.req.Path, cdh.user, cdh.proj.ID, parent.ID)
	ddir, err := cdh.dai.Dir.Insert(&datadir)
	if err != nil {
		return ddir, err
	}

	// Add the directory to the project.
	if err := cdh.dai.Project.AddDirectories(cdh.proj, ddir.ID); err != nil {
		return ddir, err
	}

	return ddir, nil
}

// getParent retrieves the parent directory for a directory path. It does
// this by getting the parent in the path name and then querying the database
// by name for this particular entry. The query is filtered by the project
// which prevents any collisions since a project is a rooted tree.
func (cdh *createDirHandler) getParent() (*schema.Directory, error) {
	var (
		parent *schema.Directory
		err    error
	)
	parentPath := filepath.Dir(cdh.req.Path)
	if parent, err = cdh.dai.Dir.ByPath(parentPath, cdh.req.ProjectID); err != nil {
		return nil, mcerr.ErrNotFound
	}
	return parent, nil
}

// validDirPath verifies that the directory path starts with the project name.
// It handles both Linux (/) and Windows (\) style slashes.
func validDirPath(projName, dirPath string) bool {
	slash := strings.Index(dirPath, "/")
	if slash == -1 {
		slash = strings.Index(dirPath, "\\")
	}
	switch {
	case slash == -1:
		return false
	case projName != dirPath[:slash]:
		return false
	default:
		return true
	}
}
