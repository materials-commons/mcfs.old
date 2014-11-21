package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/base/db"
	"github.com/materials-commons/mcfs/mcd"
)

var session *r.Session

func init() {
	mcd.InitRethinkDB()
	session, _ = db.RSession()
}
