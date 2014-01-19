package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/contrib/schema"
	"github.com/materials-commons/mcfs/protocol"
)

type lookupHandler struct {
	session *r.Session
	user    string
}

func (h *ReqHandler) lookup(req *protocol.LookupReq) (interface{}, error) {
	l := &lookupHandler{
		session: h.session,
		user:    h.user,
	}

	switch req.Type {
	case "project":
		if req.Field != "id" {
			return nil, fmt.Errorf("Projects can only be queried by id")
		}
		rql := l.projectRql(req)
		var proj schema.Project
		return l.execute(rql, &proj)

	case "datafile":
		rql := l.dataFileRql(req)
		var datafile schema.DataFile
		return l.execute(rql, &datafile)

	case "datadir":
		rql := l.dataDirRql(req)
		var datadir schema.DataDir
		return l.execute(rql, &datadir)

	default:
		return nil, fmt.Errorf("Unknown entry type %s", req.Type)
	}
}

func (l *lookupHandler) projectRql(req *protocol.LookupReq) r.RqlTerm {
	return r.Table("projects").Get(req.Value)
}

func (l *lookupHandler) dataFileRql(req *protocol.LookupReq) r.RqlTerm {
	switch req.Field {
	case "id":
		return r.Table("datafiles").Get(req.Value)
	default:
		return r.Table("datadirs").GetAllByIndex("id", req.LimitToID).
			OuterJoin(r.Table("datafiles"),
			func(ddirRow, dfRow r.RqlTerm) r.RqlTerm {
				return ddirRow.Field("datafiles").Contains(dfRow.Field("id"))
			}).Zip().Filter(r.Row.Field(req.Field).Eq(req.Value))
	}
}

func (l *lookupHandler) dataDirRql(req *protocol.LookupReq) r.RqlTerm {
	switch req.Field {
	case "id":
		return r.Table("datadirs").Get(req.Value)
	default:
		return r.Table("project2datadir").GetAllByIndex("project_id", req.LimitToID).
			EqJoin("datadir_id", r.Table("datadirs")).Zip().
			Filter(r.Row.Field(req.Field).Eq(req.Value))
	}
}

func (l *lookupHandler) execute(query r.RqlTerm, v interface{}) (interface{}, error) {
	err := model.GetRow(query, l.session, v)
	switch {
	case err != nil:
		return nil, err
	case !l.hasAccess(v):
		return nil, fmt.Errorf("Permission denied")
	default:
		return v, nil
	}
}

func (l *lookupHandler) hasAccess(v interface{}) bool {
	var owner string
	switch t := v.(type) {
	case *schema.Project:
		owner = t.Owner
	case *schema.DataDir:
		owner = t.Owner
	case *schema.DataFile:
		owner = t.Owner
	}
	return OwnerGaveAccessTo(owner, l.user, l.session)
}
