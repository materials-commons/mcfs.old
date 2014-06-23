package service

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/base/db"
	"github.com/materials-commons/mcfs/server"
)

var session *r.Session

func init() {
	mcfs.InitRethinkDB()
	session, _ = db.RSession()
}
