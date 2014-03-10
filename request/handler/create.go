package handler

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"path/filepath"
	"strings"
)

// NewCreateProject creates a new CreateProjectHandler.
func NewCreateProject(db interface{}) CreateProjectHandler {
	switch t := db.(type) {
	case *r.Session:
		return newRethinkCreateProjectHandler(t)
	default:
		return newSqlCreateProjectHandler()
	}
}

type rethinkCreateProjectHandler struct {
	session *r.Session
}

func newRethinkCreateProjectHandler(session *r.Session) CreateProjectHandler {
	return &rethinkCreateProjectHandler{
		session: session,
	}
}

func (h *rethinkCreateProjectHandler) Validate(req *protocol.CreateProjectReq) bool {
	return validateProject(req)
}

func validateProject(req *protocol.CreateProjectReq) bool {
	i := strings.Index(req.Name, "/")
	return i == -1
}

func (h *rethinkCreateProjectHandler) GetProject(name, user string) (*schema.Project, error) {
	rql := r.Table("projects").GetAllByIndex("owner", user).
		Filter(r.Row.Field("name").Eq(name))
	var project schema.Project
	err := model.GetRow(rql, h.session, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (h *rethinkCreateProjectHandler) CreateProject(name, user string) (*schema.Project, error) {
	datadir := schema.NewDataDir(name, "private", user, "")
	rv, err := r.Table("datadirs").Insert(datadir).RunWrite(h.session)
	if err != nil {
		return nil, err
	} else if rv.Inserted == 0 {
		return nil, fmt.Errorf("unable to create datadir for project")
	}
	datadirID := rv.GeneratedKeys[0]
	project := schema.NewProject(name, datadirID, user)
	rv, err = r.Table("projects").Insert(project).RunWrite(h.session)
	if err != nil {
		return nil, err
	}
	projectID := rv.GeneratedKeys[0]
	project.ID = projectID
	p2d := schema.Project2DataDir{
		ProjectID: projectID,
		DataDirID: datadirID,
	}

	// TODO: What if we get an error here?
	rv, err = r.Table("project2datadir").Insert(p2d).RunWrite(h.session)
	return &project, nil
}

type sqlCreateProjectHandler struct {
}

func newSqlCreateProjectHandler() CreateProjectHandler {
	return &sqlCreateProjectHandler{}
}

func (h *sqlCreateProjectHandler) Validate(req *protocol.CreateProjectReq) bool {
	return validateProject(req)
}

func (h *sqlCreateProjectHandler) GetProject(name, user string) (*schema.Project, error) {
	return nil, nil
}

func (h *sqlCreateProjectHandler) CreateProject(name, user string) (*schema.Project, error) {
	return nil, nil
}

type rethinkCreateDirHandler struct {
	session *r.Session
}

// NewCreateDir creates a new CreateDirHandler.
func NewCreateDir(db interface{}) CreateDirHandler {
	switch t := db.(type) {
	case *r.Session:
		return newRethinkCreateDirHandler(t)
	default:
		return newSqlCreateDirHandler()
	}
}

func newRethinkCreateDirHandler(session *r.Session) CreateDirHandler {
	return &rethinkCreateDirHandler{
		session: session,
	}
}

func (h *rethinkCreateDirHandler) GetProject(id string) (*schema.Project, error) {
	return model.GetProject(id, h.session)
}

func (h *rethinkCreateDirHandler) GetDataDir(req *protocol.CreateDirReq) (*schema.DataDir, error) {
	rql := r.Table("project2datadir").GetAllByIndex("project_id", req.ProjectID).
		EqJoin("datadir_id", r.Table("datadirs")).Zip().Filter(r.Row.Field("name").Eq(req.Path))
	var dataDir schema.DataDir
	err := model.GetRow(rql, h.session, &dataDir)
	if err != nil {
		return nil, err
	}
	return &dataDir, nil
}

func (h *rethinkCreateDirHandler) GetParent(path string) (*schema.DataDir, error) {
	parent := filepath.Dir(path)
	query := r.Table("datadirs").GetAllByIndex("name", parent)
	var d schema.DataDir
	err := model.GetRow(query, h.session, &d)
	if err != nil {
		return nil, fmt.Errorf("no parent for %s", path)
	}
	return &d, nil
}

func (h *rethinkCreateDirHandler) CreateDir(req *protocol.CreateDirReq, user, parentID string) (*schema.DataDir, error) {
	datadir := schema.NewDataDir(req.Path, "private", user, parentID)
	var wr r.WriteResponse
	wr, err := r.Table("datadirs").Insert(datadir).RunWrite(h.session)
	if err == nil && wr.Inserted > 0 {
		// Successful insert into the database
		dataDirID := wr.GeneratedKeys[0]
		p2d := schema.Project2DataDir{
			ProjectID: req.ProjectID,
			DataDirID: dataDirID,
		}
		r.Table("project2datadir").Insert(p2d).RunWrite(h.session)
		datadir.ID = dataDirID
		h.denormInsert(&datadir)
		return &datadir, nil
	}
	return nil, fmt.Errorf("unable to insert into database")
}

func (h *rethinkCreateDirHandler) denormInsert(datadir *schema.DataDir) error {
	dataDirDenorm := schema.DataDirDenorm{
		ID:        datadir.ID,
		Name:      datadir.Name,
		Owner:     datadir.Owner,
		Birthtime: datadir.Birthtime,
	}
	r.Table("datadirs_denorm").Insert(dataDirDenorm).RunWrite(h.session)
	return nil
}

func newSqlCreateDirHandler() CreateDirHandler {
	return nil
}

type rethinkCreateFileHandler struct {
	session *r.Session
}

func (h *rethinkCreateFileHandler) Validate(req *protocol.CreateFileReq) error {
	return nil
}

func (h *rethinkCreateFileHandler) CreateFile(req *protocol.CreateFileReq, user string) (*schema.DataDir, error) {
	return nil, nil
}

type sqlCreateFileHandler struct {
}

// Validate validates a CreateFileReq.
func Validate(req *protocol.CreateFileReq) error {
	return nil
}

// CreateFile creates a file.
func CreateFile(req *protocol.CreateFileReq, user string) (*schema.DataDir, error) {
	return nil, nil
}
