package rethinkdb

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/common"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/interfaces/db/model"
	"github.com/materials-commons/mcfs/mcerr"
)

type rNotes struct {
	session *r.Session
}

func NewRNotes(session *r.Session) rNotes {
	return rNotes{
		session: session,
	}
}

func (n rNotes) Insert(t common.Type, id string, note schema.Note) (*schema.Note, error) {
	note := map[string]interface{}{
		"notes": r.Row(field).Append(note),
	}
	switch t {
	case common.ProjectType:
		return nil, model.Projects.Qs(p.session).Update(id, note)
	case common.ReviewType:
		return nil, nil
	case common.SampleType:
		return nil, nil
	case common.DirectoryType:
		return nil, model.Dirs.Qs(p.session).Update(id, note)
	case common.FileType:
		return nil, model.Files.Qs(p.session).Update(id, note)
	default:
		return nil, mcerr.ErrInvalid
	}
}

func (n rNotes) Update(t common.Type, id string, note schema.Note) (*schema.Note, error) {

}

func (n rNotes) Remove(t common.Type, id string, noteID string) (*schema.Note, error) {

}
