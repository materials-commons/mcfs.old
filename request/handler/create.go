package handler

import (
	"fmt"
	"github.com/materials-commons/base/model"
	"strings"
)

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
)

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
		return nil, fmt.Errorf("Unable to create datadir for project")
	}
	datadirID := rv.GeneratedKeys[0]
	project := schema.NewProject(name, datadirID, user)
	rv, err = r.Table("projects").Insert(project).RunWrite(h.session)
	if err != nil {
		return nil, err
	}
	projectID := rv.GeneratedKeys[0]
	project.Id = projectID
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
