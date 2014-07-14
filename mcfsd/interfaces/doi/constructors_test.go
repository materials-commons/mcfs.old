package doi

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/interfaces/db"
	"github.com/materials-commons/mcfs/mcfsd"
)

var session *r.Session

func init() {
	mcfs.InitRethinkDB()
	session, _ = db.RSession()
}
