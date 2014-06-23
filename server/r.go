package mcfs

import (
	"fmt"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcfs/base/db"
)

func InitRethinkDB() {
	dbConn := config.GetString("MCDB_CONNECTION")
	dbName := config.GetString("MCDB_NAME")
	fmt.Println("InitRethinkDB:", dbConn, dbName)
	db.SetAddress(dbConn)
	db.SetDatabase(dbName)
}
