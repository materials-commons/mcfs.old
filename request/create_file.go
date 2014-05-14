package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/service"
)

type createFileHandler2 struct {
	user     string
	dirs     service.Dirs
	projects service.Projects
	files    service.Files
}

func (h *ReqHandler) createFile2(req *protocol.CreateFileReq) (resp *protocol.CreateResp, err error) {
	cfh := newCreateFileHandler2(h.user)

	// Make sure we have a valid request.
	if err := cfh.validateRequest(req, h.session); err != nil {
		return nil, err
	}

	// Check if the file already exists.
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

	default:
		// At this point the file exists and is fully uploaded. So we can create a new file
		// to upload to. We also need to update all entries entries to point to this new file.
		// New file has parent set to the old file.
		return cfh.createNewFileVersion(f, req)
	}
}

func newCreateFileHandler2(user string) *createFileHandler2 {
	return &createFileHandler2{
		user:     user,
		dirs:     service.NewDirs(service.RethinkDB),
		projects: service.NewProjects(service.RethinkDB),
		files:    service.NewFiles(service.RethinkDB),
	}
}

func (cfh *createFileHandler2) validateRequest(req *protocol.CreateFileReq, session *r.Session) error {
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

func (cfh *createFileHandler2) createNewFile(req *protocol.CreateFileReq) (*protocol.CreateResp, error) {
	file := newFile(req, cfh.user)
	dup, err := cfh.files.ByChecksum(file.Checksum)
	if err == nil {
		// Found a matching entry, set usesid to it
		file.UsesID = dup.ID
	}
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

func newFile(req *protocol.CreateFileReq, user string) *schema.File {
	file := schema.NewFile(req.Name, user)
	file.DataDirs = append(file.DataDirs, req.DataDirID)
	file.Checksum = req.Checksum
	file.Size = req.Size
	return &file
}

func (cfh *createFileHandler2) partiallyUploaded(file *schema.File, mcdir string) bool {
	id := datafileLocationID(file)
	dfSize := datafileSize(mcdir, id)

	// If the expected size of the file in the database doesn't match
	// the size of the file on disk, then the file has not been
	// completely uploaded.
	return dfSize != file.Size
}

func (cfh *createFileHandler2) createNewFileVersion(file *schema.File, req *protocol.CreateFileReq) (*protocol.CreateResp, error) {
	f := newFile(req, cfh.user)
	f.Parent = file.ID
	dup, err := cfh.files.ByChecksum(file.Checksum)
	if err == nil {
		// Found a matching entry, set usesid to it
		f.UsesID = dup.ID
	}

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

type createFileHandler struct {
	modelValidator
}

func (h *ReqHandler) createFile(req *protocol.CreateFileReq) (resp *protocol.CreateResp, err error) {
	cfh := createFileHandler{
		modelValidator: newModelValidator(h.user, h.session),
	}

	if err := cfh.validCreateFileReq(req); err != nil {
		return nil, err
	}

	df := schema.NewFile(req.Name, h.user)
	df.DataDirs = append(df.DataDirs, req.DataDirID)
	df.Checksum = req.Checksum
	df.Size = req.Size
	otherID, err := cfh.duplicateFileID(req.Checksum, req.Size)
	if err == nil && otherID != "" {
		df.UsesID = otherID
	}
	rv, err := r.Table("datafiles").Insert(df).RunWrite(h.session)
	if err != nil {
		return nil, err
	}

	if rv.Inserted == 0 {
		return nil, mc.ErrCreate
	}
	datafileID := rv.GeneratedKeys[0]

	// TODO: Eliminate an extra query to look up the DataDir
	// when we just did during verification.
	datadir, _ := model.GetDirectory(req.DataDirID, h.session)
	datadir.DataFiles = append(datadir.DataFiles, datafileID)

	// TODO: Really should check for errors here. What do
	// we do? The database could get out of sync. Maybe
	// need a way to update partially completed items by
	// putting into a log? Ugh...
	r.Table("datadirs").Update(datadir).RunWrite(h.session)
	createResp := protocol.CreateResp{
		ID: datafileID,
	}
	return &createResp, nil
}

func (h createFileHandler) validCreateFileReq(fileReq *protocol.CreateFileReq) error {
	proj, err := model.GetProject(fileReq.ProjectID, h.session)
	if err != nil {
		return fmt.Errorf("unknown project id %s", fileReq.ProjectID)
	}

	if proj.Owner != h.user {
		return fmt.Errorf("user %s is not owner of project %s", h.user, proj.Name)
	}

	datadir, err := model.GetDirectory(fileReq.DataDirID, h.session)
	if err != nil {
		return fmt.Errorf("unknown datadir Id %s", fileReq.DataDirID)
	}

	if !h.datadirInProject(datadir.ID, proj.ID) {
		return fmt.Errorf("datadir %s not in project %s", datadir.Name, proj.Name)
	}

	if h.datafileExistsInDataDir(fileReq.DataDirID, fileReq.Name) {
		return mc.ErrExists
	}

	if fileReq.Size < 1 {
		return fmt.Errorf("invalid size (%d) for datafile %s", fileReq.Size, fileReq.Name)
	}

	if fileReq.Checksum == "" {
		return fmt.Errorf("bad checksum (%s) for datafile %s", fileReq.Checksum, fileReq.Name)
	}

	return nil
}

func (h *createFileHandler) duplicateFileID(checksum string, size int64) (id string, err error) {
	rql := r.Table("datafiles").GetAllByIndex("checksum", checksum)
	var datafiles []schema.File
	err = model.GetRows(rql, h.session, &datafiles)
	if err != nil {
		return "", nil
	}

	for _, datafile := range datafiles {
		if datafile.Size == size {
			switch {
			case datafile.UsesID == "":
				return datafile.ID, nil
			default:
				return datafile.UsesID, nil
			}
		}
	}
	return "", nil
}
