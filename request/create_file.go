package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
)

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

	df := schema.NewFile(req.Name, "private", h.user)
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
