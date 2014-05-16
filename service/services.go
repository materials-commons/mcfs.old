package service

import (
	"fmt"
	"github.com/materials-commons/base/db"
	"os"
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
	dbType := getenv("MCDB_TYPE", "rethinkdb")
	dbPort := getenv("MCDB_PORT", "30815")
	dbName := getenv("MCDB_NAME", "materialscommons")

	switch strings.ToLower(dbType) {
	case "rethinkdb":
		conn := fmt.Sprintf("localhost:%s", dbPort)
		db.SetAddress(conn)
		db.SetDatabase(dbName)
		File = NewFiles(RethinkDB)
		Dir = NewDirs(RethinkDB)
		Project = NewProjects(RethinkDB)
		Group = NewGroups(RethinkDB)
		User = NewUsers(RethinkDB)
	default:
		panic(fmt.Sprintf("Unsupported database type: %s", dbType))
	}
}

func getenv(what, defaultValue string) string {
	val := os.Getenv(what)
	if val == "" {
		return defaultValue
	}

	return val
}
