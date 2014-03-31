package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/dir"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/service"
)

var _ = fmt.Println

type statProjectHandler struct {
	session *r.Session
	user    string
	files   []*dir.FileInfo
}

func (h *ReqHandler) statProject(req *protocol.StatProjectReq) (*protocol.StatProjectResp, error) {
	p := &statProjectHandler{
		session: h.session,
		user:    h.user,
	}

	var projectID string
	switch {
	case req.Name != "":
		project, err := p.getProjectByName(req.Name)
		if err != nil {
			return nil, mc.Errorm(mc.ErrNotFound, err)
		}
		projectID = project.ID
	case req.ID != "":
		projectID = req.ID
	default:
		return nil, mc.Errorm(mc.ErrInvalid, nil)
	}

	entries, err := p.projectDirList(projectID)
	if err != nil {
		return nil, mc.Errorm(mc.ErrNotFound, err)
	}

	resp := protocol.StatProjectResp{
		ProjectID: projectID,
		Entries:   entries,
	}
	return &resp, nil
}

func (p *statProjectHandler) getProjectByName(name string) (*schema.Project, error) {
	rql := r.Table("projects").GetAllByIndex("owner", p.user).Filter(r.Row.Field("name").Eq(name))
	var project schema.Project
	err := model.GetRow(rql, p.session, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (p *statProjectHandler) projectDirList(projectID string) ([]*dir.FileInfo, error) {
	projects := service.NewProjects(service.RethinkDB)
	files, err := projects.Files(projectID)
	return files, err
}
