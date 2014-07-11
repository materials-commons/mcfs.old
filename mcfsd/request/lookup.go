package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/db"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/model"
	"github.com/materials-commons/mcfs/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/server/service"
)

type lookupHandler struct {
	session *r.Session
	user    string
	service *service.Service
}

func (h *ReqHandler) lookup(req *protocol.LookupReq) (interface{}, error) {
	session, _ := db.RSession()
	l := &lookupHandler{
		session: session,
		user:    h.user,
		service: h.service,
	}

	switch req.Type {
	case "project":
		rql := l.projectRql(req)
		var proj schema.Project
		return l.execute(rql, &proj)

	case "datafile":
		rql := l.dataFileRql(req)
		var datafile schema.File
		return l.execute(rql, &datafile)

	case "datadir":
		rql := l.dataDirRql(req)
		var datadir schema.Directory
		return l.execute(rql, &datadir)

	default:
		return nil, mcerr.Errorf(mcerr.ErrInvalid, "Unknown entry type %s", req.Type)
	}
}

func (l *lookupHandler) projectRql(req *protocol.LookupReq) r.RqlTerm {
	switch req.Field {
	case "id":
		return r.Table("projects").Get(req.Value)
	default:
		return r.Table("projects").GetAllByIndex("owner", l.user).Filter(r.Row.Field(req.Field).Eq(req.Value))
	}
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
		return nil, mcerr.Errorm(mcerr.ErrInvalid, err)
	case !l.hasAccess(v):
		return nil, mcerr.Errorf(mcerr.ErrNoAccess, "Permission denied")
	default:
		return v, nil
	}
}

func (l *lookupHandler) hasAccess(v interface{}) bool {
	var owner string
	switch t := v.(type) {
	case *schema.Project:
		owner = t.Owner
	case *schema.Directory:
		owner = t.Owner
	case *schema.File:
		owner = t.Owner
	}
	return l.service.Group.HasAccess(owner, l.user)
}
