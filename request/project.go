package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/contrib/mc"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/contrib/schema"
	"github.com/materials-commons/mcfs/protocol"
)

var _ = fmt.Println

type projectEntryHandler struct {
	session *r.Session
	user    string
}

func (h *ReqHandler) projectEntries(req *protocol.ProjectEntriesReq) (*protocol.ProjectEntriesResp, error) {
	p := &projectEntryHandler{
		session: h.session,
		user:    h.user,
	}

	project, err := p.getProjectByName(req.ProjectName)
	if err != nil {
		return nil, err
	}

	entries, err := p.getProjectEntries(project.Id)
	if err != nil {
		return nil, err
	}

	resp := protocol.ProjectEntriesResp{
		ProjectID: project.Id,
		Entries:   entries,
	}
	return &resp, nil
}

func (p *projectEntryHandler) getProjectByName(name string) (*schema.Project, error) {
	rql := r.Table("projects").GetAllByIndex("owner", p.user).Filter(r.Row.Field("name").Eq(name))
	var project schema.Project
	err := model.GetRow(rql, p.session, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (p *projectEntryHandler) getProjectEntries(projectID string) ([]protocol.ProjectEntry, error) {
	rql := p.entriesRql(projectID)
	rows, err := rql.Run(p.session)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []protocol.ProjectEntry
	for rows.Next() {
		var projectEntry protocol.ProjectEntry
		err := rows.Scan(&projectEntry)
		if err != nil {
			fmt.Println("err on scan =", err)
			continue
		}
		results = append(results, projectEntry)
	}

	if len(results) == 0 {
		return nil, mc.ErrNotFound
	}

	return results, nil
}

var dataDirMergeMap = map[string]interface{}{
	"datadir_name": r.Row.Field("name"),
	"datadir_id":   r.Row.Field("id"),
}

func (p *projectEntryHandler) entriesRql(projectID string) r.RqlTerm {
	return r.Table("project2datadir").GetAllByIndex("project_id", projectID).
		EqJoin("datadir_id", r.Table("datadirs")).Zip().Map(r.Row.Merge(dataDirMergeMap)).
		Without("name", "id").OrderBy("datadir_name").
		OuterJoin(r.Table("datafiles"),
		func(ddirRow, dfRow r.RqlTerm) r.RqlTerm {
			return ddirRow.Field("datafiles").Contains(dfRow.Field("id"))
		}).Zip().Pluck("datadir_name", "datadir_id",
		"name", "id", "size", "checksum")
}
