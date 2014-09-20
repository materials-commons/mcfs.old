package mcfs

import (
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/db"
)

func InitRethinkDB() {
	dbConn := config.GetString("MCDB_CONNECTION")
	dbName := config.GetString("MCDB_NAME")
	db.SetAddress(dbConn)
	db.SetDatabase(dbName)
}
