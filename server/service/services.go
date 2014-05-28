package service

import (
	"fmt"
	"github.com/materials-commons/mcfs/base/db"
	"github.com/materials-commons/gohandy/env"
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

// Setup the services in the init so that they can be configured to
// the type of database to connect to.
func init() {
	dbType := env.GetDefault("MCDB_TYPE", "rethinkdb")

	switch strings.ToLower(dbType) {
	case "rethinkdb":
		setupRethinkDB()
	default:
		panic(fmt.Sprintf("Unsupported database type: %s", dbType))
	}
}

// setupRethinkDB connects the services to rethinkdb. In a production setting
// the environment variables should always be set. The defaults are set to
// the test environment.
func setupRethinkDB() {
	dbPort := env.GetDefault("MCDB_PORT", "30815")
	dbName := env.GetDefault("MCDB_NAME", "materialscommons")
	conn := fmt.Sprintf("localhost:%s", dbPort)

	db.SetAddress(conn)
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
