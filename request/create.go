package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request/handler"
	"path/filepath"
	"strings"
)

func (h *ReqHandler) createProject(req *protocol.CreateProjectReq) (resp *protocol.CreateProjectResp, s *stateStatus) {
	projHandler := handler.NewCreateProject(h.session)

	if !projHandler.Validate(req) {
		s = ssf(mc.ErrorCodeInvalid, "Invalid project name %s", req.Name)
		return nil, s
	}

	proj, err := projHandler.GetProject(req.Name, h.user)
	switch {
	case err == nil:
		// Found project
		resp := &protocol.CreateProjectResp{
			ProjectID: proj.Id,
			DataDirID: proj.DataDir,
		}
		return resp, ss(mc.ErrorCodeExists, mc.ErrExists)

	default:
		p, err := projHandler.CreateProject(req.Name, h.user)
		if err != nil {
			s.status = mc.ErrorCodeCreate
			s.err = err
			return nil, s
		}
		resp := &protocol.CreateProjectResp{
			ProjectID: p.Id,
			DataDirID: p.DataDir,
		}
		return resp, nil
	}
}

type createFileHandler struct {
	modelValidator
}

func (h *ReqHandler) createFile(req *protocol.CreateFileReq) (resp *protocol.CreateResp, s *stateStatus) {
	cfh := createFileHandler{
		modelValidator: newModelValidator(h.user, h.session),
	}

	if err := cfh.validCreateFileReq(req); err != nil {
		if err == mc.ErrExists {
			s = ss(mc.ErrorCodeExists, err)
		} else {
			s = ss(mc.ErrorCodeInvalid, err)
		}
		return nil, s
	}

	df := schema.NewDataFile(req.Name, "private", h.user)
	df.DataDirs = append(df.DataDirs, req.DataDirID)
	df.Checksum = req.Checksum
	df.Size = req.Size
	otherId, err := cfh.duplicateFileId(req.Checksum, req.Size)
	if err == nil && otherId != "" {
		df.UsesID = otherId
	}
	rv, err := r.Table("datafiles").Insert(df).RunWrite(h.session)
	if err != nil {
		s = ss(mc.ErrorCodeCreate, err)
		return nil, s
	}

	if rv.Inserted == 0 {
		s = ssf(mc.ErrorCodeCreate, "Unable to insert datafile")
		return nil, s
	}
	datafileId := rv.GeneratedKeys[0]

	// TODO: Eliminate an extra query to look up the DataDir
	// when we just did during verification.
	datadir, _ := model.GetDataDir(req.DataDirID, h.session)
	datadir.DataFiles = append(datadir.DataFiles, datafileId)

	// TODO: Really should check for errors here. What do
	// we do? The database could get out of sync. Maybe
	// need a way to update partially completed items by
	// putting into a log? Ugh...
	r.Table("datadirs").Update(datadir).RunWrite(h.session)
	createResp := protocol.CreateResp{
		ID: datafileId,
	}
	return &createResp, nil
}

func (h createFileHandler) validCreateFileReq(fileReq *protocol.CreateFileReq) error {
	proj, err := model.GetProject(fileReq.ProjectID, h.session)
	if err != nil {
		return fmt.Errorf("Unknown project id %s", fileReq.ProjectID)
	}

	if proj.Owner != h.user {
		return fmt.Errorf("User %s is not owner of project %s", h.user, proj.Name)
	}

	datadir, err := model.GetDataDir(fileReq.DataDirID, h.session)
	if err != nil {
		return fmt.Errorf("Unknown datadir Id %s", fileReq.DataDirID)
	}

	if !h.datadirInProject(datadir.Id, proj.Id) {
		return fmt.Errorf("Datadir %s not in project %s", datadir.Name, proj.Name)
	}

	if h.datafileExistsInDataDir(fileReq.DataDirID, fileReq.Name) {
		return mc.ErrExists
	}

	if fileReq.Size < 1 {
		return fmt.Errorf("Invalid size (%d) for datafile %s", fileReq.Size, fileReq.Name)
	}

	if fileReq.Checksum == "" {
		return fmt.Errorf("Bad checksum (%s) for datafile %s", fileReq.Checksum, fileReq.Name)
	}

	return nil
}

func (h *createFileHandler) duplicateFileId(checksum string, size int64) (id string, err error) {
	rql := r.Table("datafiles").GetAllByIndex("checksum", checksum)
	var datafiles []schema.DataFile
	err = model.GetRows(rql, h.session, &datafiles)
	if err != nil {
		return "", nil
	}

	for _, datafile := range datafiles {
		if datafile.Size == size {
			switch {
			case datafile.UsesID == "":
				return datafile.Id, nil
			default:
				return datafile.UsesID, nil
			}
		}
	}
	return "", nil
}

func (h *ReqHandler) createDir(req *protocol.CreateDirReq) (resp *protocol.CreateResp, s *stateStatus) {
	v := newModelValidator(h.user, h.session)
	if v.verifyProject(req.ProjectID) {
		return h.createDataDir(req)
	}
	return nil, ssf(mc.ErrorCodeInvalid, "Invalid project: %s", req.ProjectID)
}

func (h *ReqHandler) createDataDir(req *protocol.CreateDirReq) (resp *protocol.CreateResp, s *stateStatus) {
	dh := handler.NewCreateDir(h.session)
	var datadir schema.DataDir
	proj, err := dh.GetProject(req.ProjectID)
	switch {
	case err != nil:
		return nil, ssf(mc.ErrorCodeInvalid, "Bad projectID %s", req.ProjectID)
	case proj.Owner != h.user:
		return nil, ssf(mc.ErrorCodeNoAccess, "Access to project not allowed")
	case !validDirPath(proj.Name, req.Path):
		return nil, ssf(mc.ErrorCodeInvalid, "Invalid directory path %s", req.Path)
	default:
		id, err := dataDirIdForName(req.ProjectID, req.Path, h.session)
		switch {
		case err == mc.ErrNotFound:
			var parent string
			if parent, err = getParent(req.Path, h.session); err != nil {
				return nil, ss(mc.ErrorCodeNotFound, err)
			}
			datadir = schema.NewDataDir(req.Path, "private", h.user, parent)
			var wr r.WriteResponse
			wr, err = r.Table("datadirs").Insert(datadir).RunWrite(h.session)
			if err == nil && wr.Inserted > 0 {
				dataDirID := wr.GeneratedKeys[0]
				p2d := Project2Datadir{
					ProjectID: req.ProjectID,
					DataDirID: dataDirID,
				}
				r.Table("project2datadir").Insert(p2d).RunWrite(h.session)
				resp := &protocol.CreateResp{
					ID: dataDirID,
				}
				return resp, nil
			}
			// TODO: Make error message better
			return nil, ssf(mc.ErrorCodeCreate, "Unable to insert into database")
		case err != nil:
			return nil, ss(mc.ErrorCodeNotFound, err)
		default:
			resp := &protocol.CreateResp{
				ID: id,
			}
			return resp, nil
		}
	}
}

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

func getParent(ddirPath string, session *r.Session) (string, error) {
	parent := filepath.Dir(ddirPath)
	query := r.Table("datadirs").GetAllByIndex("name", parent)
	var d schema.DataDir
	err := model.GetRow(query, session, &d)
	if err != nil {
		return "", fmt.Errorf("No parent for %s", ddirPath)
	}
	return d.Id, nil
}

func dataDirIdForName(projectID, path string, session *r.Session) (string, error) {
	rql := r.Table("project2datadir").GetAllByIndex("project_id", projectID).
		EqJoin("datadir_id", r.Table("datadirs")).Zip().Filter(r.Row.Field("name").Eq(path))
	var dataDir schema.DataDir
	err := model.GetRow(rql, session, &dataDir)
	if err != nil {
		return "", err
	}
	return dataDir.Id, nil
}
