package request

import (
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/server/service"
)

// createFileHandler is an internal type for handling create file requests.
type createFileHandler struct {
	user    string
	service *service.Service
}

// createFile will create a new file, or use an existing file. Existing files are
// returned if the existing files upload was interrupted. In the case where an
// existing file is returned, the checksums must match between the request and
// the existing file.
func (h *ReqHandler) createFile(req *protocol.CreateFileReq) (resp *protocol.CreateResp, err error) {
	cfh := newCreateFileHandler(h.user, h.service)

	if err := cfh.validateRequest(req); err != nil {
		return nil, err
	}

	// Check the file status.
	files, err := h.service.File.ByPathChecksum(req.Name, req.DataDirID, req.Checksum)
	switch {
	case len(files) == 0:
		// This is the easy case. No matching files were found, so we just create a
		// new file. There may be an existing current file with a different checksum
		// so we need to handle that case as well.
		return cfh.createNewFile(req)

	case len(files) == 1:
		// Only one match. We have either a partial, or a fully uploaded file.
		// Either way it doesn't matter we just let the upload state figure out
		// what to do.
		f := files[0]
		return &protocol.CreateResp{ID: f.ID}, nil

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
			return &protocol.CreateResp{ID: partial.ID}, nil

		case current != nil:
			// The checksum matches the current file. Return that and let
			// the uploader take care of things.
			return &protocol.CreateResp{ID: current.ID}, nil

		default:
			// No partial, and not match on existing. Create a new
			// file entry to write to.
			return cfh.createNewFile(req)
		}
	}
}

func newCreateFileHandler(user string, service *service.Service) *createFileHandler {
	return &createFileHandler{
		user:    user,
		service: service,
	}
}

// validateRequest will validate the CreateFileReq. It does sanity checking on the file
// size and checksum. We rely on the client to send us a good checksum.
func (cfh *createFileHandler) validateRequest(req *protocol.CreateFileReq) error {
	proj, err := cfh.service.Project.ByID(req.ProjectID)
	if err != nil {
		return mcerr.Errorf(mcerr.ErrInvalid, "Bad projectID %s", req.ProjectID)
	}

	if !cfh.service.Group.HasAccess(proj.Owner, cfh.user) {
		return mcerr.ErrNoAccess
	}

	ddir, err := cfh.service.Dir.ByID(req.DataDirID)
	if err != nil {
		return mcerr.Errorf(mcerr.ErrInvalid, "Unknown directory id: %s", req.DataDirID)
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
func (cfh *createFileHandler) createNewFile(req *protocol.CreateFileReq) (*protocol.CreateResp, error) {
	var f *schema.File
	currentFile, err := cfh.service.File.ByPath(req.Name, req.DataDirID)
	switch {
	case err == mcerr.ErrNotFound:
		// There is no current entry, create a new one.
		f = cfh.newFile(req)
	case err != nil:
		// Database error occurred.
		return nil, err
	default:
		// There is a current entry, so create the new one with the parent pointing
		// to the current entry.
		f = cfh.newFile(req)
		f.Parent = currentFile.ID
	}

	created, err := cfh.service.File.InsertEntry(f)
	if err != nil {
		return nil, err
	}

	return &protocol.CreateResp{ID: created.ID}, nil
}

// newFile creates a new file object to insert into the database. It also handles the
// bookkeeping task of setting the usesid field if the upload is for a previously
// uploaded file.
func (cfh *createFileHandler) newFile(req *protocol.CreateFileReq) *schema.File {
	file := schema.NewFile(req.Name, cfh.user)
	file.DataDirs = append(file.DataDirs, req.DataDirID)
	file.Checksum = req.Checksum
	file.Size = req.Size
	file.Current = false

	dup, err := cfh.service.File.ByChecksum(file.Checksum)
	if err == nil && dup != nil {
		// Found a matching entry, set usesid to it
		file.UsesID = dup.ID
	}

	return &file
}
