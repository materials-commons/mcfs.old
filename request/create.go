package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/contrib/mc"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/contrib/schema"
	"github.com/materials-commons/mcfs/protocol"
	"path/filepath"
	"strings"
)

func (h *ReqHandler) createProject(req *protocol.CreateProjectReq) (*protocol.CreateProjectResp, error) {
	switch {
	case !validProjectName(req.Name):
		return nil, fmt.Errorf("Invalid project name %s", req.Name)
	case projectExists(req.Name, h.user, h.session):
		return nil, fmt.Errorf("Project %s exists", req.Name)
	default:
		projectId, datadirId, err := projectCreate(req.Name, h.user, h.session)
		if err != nil {
			return nil, err
		}
		resp := &protocol.CreateProjectResp{
			ProjectID: projectId,
			DataDirID: datadirId,
		}
		return resp, nil
	}
}

func validProjectName(projectName string) bool {
	i := strings.Index(projectName, "/")
	return i == -1
}

func projectExists(projectName, user string, session *r.Session) bool {
	results, err := r.Table("projects").GetAllByIndex("owner", user).
		Filter(r.Row.Field("name").Eq(projectName)).
		Run(session)
	if err != nil {
		return true // Error, we don't know if it exists
	}
	defer results.Close()

	return results.Next()
}

func projectCreate(projectName, user string, session *r.Session) (projectId, datadirId string, err error) {
	datadir := schema.NewDataDir(projectName, "private", user, "")
	rv, err := r.Table("datadirs").Insert(datadir).RunWrite(session)
	if err != nil {
		return "", "", err
	}
	datadirId = datadir.Id
	project := schema.NewProject(projectName, datadirId, user)
	rv, err = r.Table("projects").Insert(project).RunWrite(session)
	if err != nil {
		return "", "", err
	}
	return rv.GeneratedKeys[0], datadirId, nil
}

type createFileValidator struct {
	modelValidator
}

func (h *ReqHandler) createFile(req *protocol.CreateFileReq) (*protocol.CreateResp, error) {
	v := createFileValidator{
		modelValidator: newModelValidator(h.user, h.session),
	}

	if err := v.validCreateFileReq(req); err != nil {
		return nil, err
	}

	df := schema.NewDataFile(req.Name, "private", h.user)
	df.DataDirs = append(df.DataDirs, req.DataDirID)
	df.Checksum = req.Checksum
	df.Size = req.Size
	rv, err := r.Table("datafiles").Insert(df).RunWrite(h.session)
	if err != nil {
		return nil, err
	}

	if rv.Inserted == 0 {
		return nil, fmt.Errorf("Unable to insert datafile")
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

func (v createFileValidator) validCreateFileReq(fileReq *protocol.CreateFileReq) error {
	proj, err := model.GetProject(fileReq.ProjectID, v.session)
	if err != nil {
		return fmt.Errorf("Unknown project id %s", fileReq.ProjectID)
	}

	if proj.Owner != v.user {
		return fmt.Errorf("User %s is not owner of project %s", v.user, proj.Name)
	}

	datadir, err := model.GetDataDir(fileReq.DataDirID, v.session)
	if err != nil {
		return fmt.Errorf("Unknown datadir Id %s", fileReq.DataDirID)
	}

	if !v.datadirInProject(datadir.Id, proj.Id) {
		return fmt.Errorf("Datadir %s not in project %s", datadir.Name, proj.Name)
	}

	if v.datafileExistsInDataDir(fileReq.DataDirID, fileReq.Name) {
		return fmt.Errorf("Datafile %s already exists in datadir %s", fileReq.Name, datadir.Name)
	}

	if fileReq.Size < 1 {
		return fmt.Errorf("Invalid size (%d) for datafile %s", fileReq.Size, fileReq.Name)
	}

	if fileReq.Checksum == "" {
		return fmt.Errorf("Bad checksum (%s) for datafile %s", fileReq.Checksum, fileReq.Name)
	}

	return nil
}

func (h *ReqHandler) createDir(req *protocol.CreateDirReq) (*protocol.CreateResp, error) {
	v := newModelValidator(h.user, h.session)
	if v.verifyProject(req.ProjectID) {
		return h.createDataDir(req)
	}
	return nil, fmt.Errorf("Invalid project: %s", req.ProjectID)
}

func (h *ReqHandler) createDataDir(req *protocol.CreateDirReq) (*protocol.CreateResp, error) {
	var datadir schema.DataDir
	proj, err := model.GetProject(req.ProjectID, h.session)
	switch {
	case err != nil:
		return nil, fmt.Errorf("Bad projectID %s", req.ProjectID)
	case proj.Owner != h.user:
		return nil, fmt.Errorf("Access to project not allowed")
	case !validDirPath(proj.Name, req.Path):
		return nil, fmt.Errorf("Invalid directory path %s", req.Path)
	default:
		id, err := dataDirIdForName(req.ProjectID, req.Path, h.session)
		switch {
		case err == mc.ErrNotFound:
			var parent string
			if parent, err = getParent(req.Path, h.session); err != nil {
				return nil, err
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
			return nil, fmt.Errorf("Unable to insert into database")
		case err != nil:
			return nil, err
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
