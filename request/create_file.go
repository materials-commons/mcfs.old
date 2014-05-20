package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/service"
)

// createFileHandler is an internal type for handling create file requests.
type createFileHandler struct {
	user string
}

// createFile will create a new file, or use an existing file. Existing files are
// returned if the existing files upload was interrupted. In the case where an
// existing file is returned, the checksums must match between the request and
// the existing file.
func (h *ReqHandler) createFile(req *protocol.CreateFileReq) (resp *protocol.CreateResp, err error) {
	cfh := newCreateFileHandler(h.user)

	if err := cfh.validateRequest(req); err != nil {
		return nil, err
	}

	// Check the file status.
	files, err := service.File.ByPathChecksum(req.Name, req.DataDirID, req.Checksum)
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
		return &protocol.CreateResp{ID: f.FileID()}, nil

	default:
		// There are multiple matches. That means we could have old versions,
		// or a partial. Lets see if there is a partial. If there is then
		// this is easy, we just return the partial. If there isn't then we
		// need to create a new file version.
		current := schema.Files.Find(files, func(f schema.File) bool { return f.Current })
		partial := schema.Files.Find(files, func(f schema.File) bool { return f.Size != f.Uploaded })
		if partial != nil {
			return &protocol.CreateResp{ID: partial.FileID()}, nil
		}

		if current != nil {
			// Matched on a current file. Just return it.
			return &protocol.CreateResp{ID: current.FileID()}, nil
		}

		return cfh.createNewFile(req)
	}
	/*
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
			return &protocol.CreateResp{ID: f.ID}, nil

		case f.Checksum == req.Checksum:
			// File exists, is fully uploaded, and the request is to upload
			// a file with the same checksum. Just return the existing file
			// and the let the upload take care of the number of bytes.
			return &protocol.CreateResp{ID: f.ID}, nil

		default:
			// At this point the file exists and is fully uploaded, and the request has a
			// different checksum. This means a new version of the file is being uploaded.
			// We will create a new file entry to upload to. We also need to update all
			// objects to point to this new file and hide the old file. To keep track of
			// the graph the new file has parent set to the old file entry it is replacing.
			return cfh.createNewFileVersion(f, req)
		}
	*/
}

func newCreateFileHandler(user string) *createFileHandler {
	return &createFileHandler{
		user: user,
	}
}

// validateRequest will validate the CreateFileReq. It does sanity checking on the file
// size and checksum. We rely on the client to send us a good checksum.
func (cfh *createFileHandler) validateRequest(req *protocol.CreateFileReq) error {
	proj, err := service.Project.ByID(req.ProjectID)
	if err != nil {
		return mc.Errorf(mc.ErrInvalid, "Bad projectID %s", req.ProjectID)
	}

	if !service.Group.HasAccess(proj.Owner, cfh.user) {
		return mc.ErrNoAccess
	}

	ddir, err := service.Dir.ByID(req.DataDirID)
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
	var f *schema.File
	currentFile, err := service.File.ByPath(req.Name, req.DataDirID)
	switch {
	case err == mc.ErrNotFound:
		// There is no current entry, just create a new one.
		f := cfh.newFile(req)
	case err != nil:
		// Database error occured.
		return nil, err
	default:
		// There is a current entry, so create the new one with a parent pointing
		// to the current entry.
		f := cfh.newFile(req)
		f.Parent = currentFile.ID
	}
	
	created, err := service.File.InsertEntry(f)
	if err != nil {
		return nil, err
	}

	return &createResp := protocol.CreateResp{ID: created.ID}, nil
}

// newFile creates a new file object to insert into the database. It also handles the
// bookkeeping task of setting the usesid field if the upload is for a previously
// uploaded file.
func (cfh *createFileHandler) newFile(req *protocol.CreateFileReq) *schema.File {
	file := schema.NewFile(req.Name, cfh.user)
	file.DataDirs = append(file.DataDirs, req.DataDirID)
	file.Checksum = req.Checksum
	file.Size = req.Size

	dup, err := service.File.ByChecksum(file.Checksum)
	if err == nil && dup != nil {
		// Found a matching entry, set usesid to it
		file.UsesID = dup.ID
	}

	return &file
}

// createNewFileVersion creates a new version of an existing file. It handles hiding the old
// version and setting the parent on the new version to point at the old file.
func (cfh *createFileHandler) createNewFileVersion(file *schema.File, req *protocol.CreateFileReq) (*protocol.CreateResp, error) {
	f := cfh.newFile(req, cfh.user)
	f.Parent = file.ID

	// Hide the old file, but keep it around so we can get to it if needed.
	service.File.Hide(file)

	created, err := service.File.Insert(f)
	if err != nil {
		return nil, err
	}

	// New file entry created
	createResp := protocol.CreateResp{
		ID: created.ID,
	}
	return &createResp, nil
}
