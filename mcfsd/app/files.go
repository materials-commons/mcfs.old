package app

import (
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
)

type File struct {
	Name        string
	ProjectID   string
	DirectoryID string
	Checksum    string
	Owner       string
	Size        int64
}

type FilesService interface {
	Create(file File) (*schema.File, error)
}

type filesService struct {
	files    dai.Files
	dirs     dai.Dirs
	projects dai.Projects
	groups   dai.Groups
}

// NewFilesService returns a new FilesService.
func NewFilesService(files dai.Files, dirs dai.Dirs, projects dai.Projects, groups dai.Groups) *filesService {
	return &filesService{
		files:    files,
		dirs:     dirs,
		projects: projects,
		groups:   groups,
	}
}

/*
//
func (s *filesService) Create(file File) (*schema.File, error) {
	var (
		req  protocol.CreateFileReq
		resp *protocol.CreateFileResp
		cerr error
	)

	if err := s.validateRequest(f); err != nil {
		return nil, err
	}

	// Check the file status.
	files, err := r.files.ByPathChecksum(req.Name, req.DirectoryID, req.Checksum)
	switch {
	case len(files) == 0:
		// This is the easy case. No matching files were found, so we just create a
		// new file. There may be an existing current file with a different checksum
		// so we need to handle that case as well.
		resp, cerr = r.createNewFile(&req, user.Name)

	case len(files) == 1:
		// Only one match. We have either a partial, or a fully uploaded file.
		// Either way it doesn't matter we just let the upload state figure out
		// what to do.
		f := files[0]
		resp = &protocol.CreateFileResp{FileID: f.ID}

	default:
		// There are multiple matches. That means we could have old versions,
		// or a partial. Lets see if there is a partial. If there is then
		// this is easy, we just return the partial. If there isn't then the
		// checksum could match the current file. Return that if true, otherwise
		// we need to create a new file version.
		current := schema.Files.Find(files, func(f schema.File) bool { return f.Current })
		partial := schema.Files.Find(files, func(f schema.File) bool { return f.Size != f.Uploaded && !f.Current })
		switch {
		case partial != nil:
			// There is an existing partial. Use that so the upload can complete.
			resp = &protocol.CreateFileResp{FileID: partial.ID}

		case current != nil:
			// The checksum matches the current file. Return that and let
			// the uploader take care of things.
			resp = &protocol.CreateFileResp{FileID: current.ID}

		default:
			// No partial, and not match on existing. Create a new
			// file entry to write to.
			resp, cerr = r.createNewFile(&req, user.Name)
		}
	}

	if cerr != nil {
		return rest.HTTPErrore(http.StatusBadRequest, err)
	}

	response.WriteEntity(resp)
	return nil
}

// validateRequest will validate the CreateFileReq. It does sanity checking on the file
// size and checksum. We rely on the client to send us a good checksum.
func (r *filesResource) validateRequest(req *protocol.CreateFileReq, user string) error {
	proj, err := r.projects.ByID(req.ProjectID)
	if err != nil {
		return mcerr.Errorf(mcerr.ErrInvalid, "Bad projectID %s", req.ProjectID)
	}

	if !r.groups.HasAccess(proj.Owner, user) {
		return mcerr.ErrNoAccess
	}

	ddir, err := r.dirs.ByID(req.DirectoryID)
	if err != nil {
		return mcerr.Errorf(mcerr.ErrInvalid, "Unknown directory id: %s", req.DirectoryID)
	}

	if ddir.Project != req.ProjectID {
		return mcerr.Errorf(mcerr.ErrInvalid, "Directory %s not in project %s", ddir.Name, req.ProjectID)
	}

	if req.Size < 1 {
		return mcerr.Errorf(mcerr.ErrInvalid, "Invalid size (%d) for file %s", req.Size, req.Name)
	}

	if req.Checksum == "" {
		return mcerr.Errorf(mcerr.ErrInvalid, "Bad checksum (%s) for file %s", req.Checksum, req.Name)
	}

	return nil
}

// createNewFile will create the file object in the database. It inserts a new file entry
// but doesn't attach it up to dependent objects. This will happen when the upload has
// completed. If we did it before we could end up with file entries that look valid but
// their backing physical file doesn't contain all the bytes.
func (r *filesResource) createNewFile(req *protocol.CreateFileReq, user string) (*protocol.CreateFileResp, error) {
	var f *schema.File

	currentFile, err := r.files.ByPath(req.Name, req.DirectoryID)
	switch {
	case err == mcerr.ErrNotFound:
		// There is no current entry, create a new one.
		f = r.newFile(req, user)
	case err != nil:
		// Database error occurred.
		return nil, err
	default:
		// There is a current entry, so create the new one with the parent pointing
		// to the current entry.
		f = r.newFile(req, user)
		f.Parent = currentFile.ID
	}

	created, err := r.files.InsertEntry(f)
	if err != nil {
		return nil, err
	}

	return &protocol.CreateFileResp{FileID: created.ID}, nil
}

// newFile creates a new file object to insert into the database. It also handles the
// bookkeeping task of setting the usesid field if the upload is for a previously
// uploaded file.
func (r *filesResource) newFile(req *protocol.CreateFileReq, user string) *schema.File {
	file := schema.NewFile(req.Name, user)
	file.DataDirs = append(file.DataDirs, req.DirectoryID)
	file.Checksum = req.Checksum
	file.Size = req.Size
	file.Current = false

	dup, err := r.files.ByChecksum(file.Checksum)
	if err == nil && dup != nil {
		// Found a matching entry, set usesid to it
		file.UsesID = dup.ID
	}

	return &file
}
*/
