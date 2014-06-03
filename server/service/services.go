package service

import (
	"fmt"
	"github.com/materials-commons/config"
	"github.com/materials-commons/mcfs/base/db"
	"strings"
)

// Global File service
var File Files

// Global Dir service
var Dir Dirs

// Global Project service
var Project Projects

// Global Group service
var Group Groups

// Global User service
var User Users

// Init sets up the services.
func Init() {
	dbType := config.GetString("MCDB_TYPE")

	switch {
	case strings.ToLower(dbType) == "rethinkdb":
		setupRethinkDB()
	default:
		panic(fmt.Sprintf("Unsupported database type: %s", dbType))
	}
}

// setupRethinkDB connects the services to rethinkdb. In a production setting
// the environment variables should always be set. The defaults are set to
// the test environment.
func setupRethinkDB() {
	dbConn := config.GetString("MCDB_CONNECTION")
	dbName := config.GetString("MCDB_NAME")

	db.SetAddress(dbConn)
	db.SetDatabase(dbName)
	connectServices(RethinkDB)
}

// connectServices takes care of instantiating each of the services
// to the correct service database. Add new services to this method.
func connectServices(serviceDatabase ServiceDatabase) {
	File = NewFiles(serviceDatabase)
	Dir = NewDirs(serviceDatabase)
	Project = NewProjects(serviceDatabase)
	Group = NewGroups(serviceDatabase)
	User = NewUsers(serviceDatabase)
}
