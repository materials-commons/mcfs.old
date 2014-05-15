package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/service"
)

// createFileHandler is an internal type for handling create file requests.
type createFileHandler struct {
	user     string
	dirs     service.Dirs
	projects service.Projects
	files    service.Files
}

// createFile will create a new file, or use an existing file. Existing files are
// returned if the existing files upload was interrupted. In the case where an
// existing file is returned, the checksums must match.
func (h *ReqHandler) createFile(req *protocol.CreateFileReq) (resp *protocol.CreateResp, err error) {
	cfh := newCreateFileHandler(h.user)

	// Make sure we have a valid request.
	if err := cfh.validateRequest(req, h.session); err != nil {
		return nil, err
	}

	// Check if the file status.
	f, err := cfh.files.ByPath(req.Name, req.DataDirID)
	switch {
	case err == mc.ErrNotFound:
		// File doesn't exist. This is the easy case: Create a new one and
		// return its id.
		return cfh.createNewFile(req)

	case cfh.partiallyUploaded(f, h.mcdir):
		// File exists and the previous upload was only partially finished.
		// Hopefully the user is asking us to complete the upload.
		if f.Checksum != req.Checksum {
			// Uh oh, they are sending us a new version of the file when
			// the previous version has not completed its upload.
			//
			// Currently this is an unrecoverable error. The situation
			// here is that we have an existing file that has not completed
			// its upload. Now we are trying to upload a new version of the
			// file. We know its a new version because the checksums don't
			// match. This is a situation we will need to deal with.
			return nil, mc.Errorf(mc.ErrInvalid, "Attempt to upload a new file version when the previous has not completed")
		}

		// If we are here then the user is uploading the remaining bits of an existing file.
		createResp := protocol.CreateResp{
			ID: f.ID,
		}
		return &createResp, nil

	case f.Checksum == req.Checksum:
		// File exists, is fully uploaded, and the request is to upload
		// a file with the same checksum. Just return the existing file
		// and the let the upload take care of the number of bytes.
		createResp := protocol.CreateResp{
			ID: f.ID,
		}
		return &createResp, nil

	default:
		// At this point the file exists and is fully uploaded. So we can create a new file
		// to upload to. We also need to update all entries entries to point to this new file.
		// New file has parent set to the old file.
		return cfh.createNewFileVersion(f, req)
	}
}

func newCreateFileHandler(user string) *createFileHandler {
	return &createFileHandler{
		user:     user,
		dirs:     service.NewDirs(service.RethinkDB),
		projects: service.NewProjects(service.RethinkDB),
		files:    service.NewFiles(service.RethinkDB),
	}
}

// validateRequest will validate the CreateFileReq. It does sanity checking on the file
// size and checksum. We rely on the client to send us a good checksum.
func (cfh *createFileHandler) validateRequest(req *protocol.CreateFileReq, session *r.Session) error {
	proj, err := cfh.projects.ByID(req.ProjectID)
	if err != nil {
		return mc.Errorf(mc.ErrInvalid, "Bad projectID %s", req.ProjectID)
	}

	if !OwnerGaveAccessTo(proj.Owner, cfh.user, session) {
		return mc.ErrNoAccess
	}

	ddir, err := cfh.dirs.ByID(req.DataDirID)
	if err != nil {
		return mc.Errorf(mc.ErrInvalid, "Unknown directory id: %s", req.DataDirID)
	}

	if ddir.Project != req.ProjectID {
		return mc.Errorf(mc.ErrInvalid, "Directory %s not in project %s", ddir.Name, req.ProjectID)
	}

	if req.Size < 1 {
		return mc.Errorf(mc.ErrInvalid, "Invalid size (%d) for file %s", req.Size, req.Name)
	}

	if req.Checksum == "" {
		return mc.Errorf(mc.ErrInvalid, "Bad checksum (%s) for file %s", req.Checksum, req.Name)
	}

	return nil
}

// createNewFile will create the file object in the database.
func (cfh *createFileHandler) createNewFile(req *protocol.CreateFileReq) (*protocol.CreateResp, error) {
	file := cfh.newFile(req, cfh.user)
	created, err := cfh.files.Insert(file)
	if err != nil {
		// Insert into database failed
		return nil, err
	}

	// New file entry created
	createResp := protocol.CreateResp{
		ID: created.ID,
	}
	return &createResp, nil
}

// newFile creates a new file object to insert into the database. It also handles the
// bookkeeping task of setting the usesid field if the upload is for a previously
// uploaded file.
func (cfh *createFileHandler) newFile(req *protocol.CreateFileReq, user string) *schema.File {
	file := schema.NewFile(req.Name, user)
	file.DataDirs = append(file.DataDirs, req.DataDirID)
	file.Checksum = req.Checksum
	file.Size = req.Size

	dup, err := cfh.files.ByChecksum(file.Checksum)
	if err == nil {
		// Found a matching entry, set usesid to it
		file.UsesID = dup.ID
	}

	return &file
}

// partiallyUploaded checks if the file request is for a file that has not completed
// its upload.
func (cfh *createFileHandler) partiallyUploaded(file *schema.File, mcdir string) bool {
	id := datafileLocationID(file)
	dfSize := datafileSize(mcdir, id)

	// If the expected size of the file in the database doesn't match
	// the size of the file on disk, then the file has not been
	// completely uploaded.
	return dfSize != file.Size
}

// createNewFileVersion creates a new version of an existing file. It handles hiding the old
// version and setting the parent on the new version to point at the old file.
func (cfh *createFileHandler) createNewFileVersion(file *schema.File, req *protocol.CreateFileReq) (*protocol.CreateResp, error) {
	f := cfh.newFile(req, cfh.user)
	f.Parent = file.ID

	// Hide the old file, but keep it around so we can get to it if needed.
	cfh.files.Hide(file)

	// Insert the new file into the database.
	created, err := cfh.files.Insert(f)
	if err != nil {
		// Insert into database failed
		return nil, err
	}

	// New file entry created
	createResp := protocol.CreateResp{
		ID: created.ID,
	}
	return &createResp, nil
}
