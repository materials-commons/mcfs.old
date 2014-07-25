package testutils

import (
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai/rethinkdb"
	"github.com/materials-commons/mcfs/testutils/tdb"
)

// NewRFiles creates a new dai.Files connected to the test database.
func NewRFiles() dai.Files {
	return rethinkdb.NewRFiles(tdb.NewSession())
}

// NewRDirs creates a new dai.Dirs connected to the test database.
func NewRDirs() dai.Dirs {
	return rethinkdb.NewRDirs(tdb.NewSession())
}

// NewRProjects creates a new dai.Projects connected to the test database.
func NewRProjects() dai.Projects {
	return rethinkdb.NewRProjects(tdb.NewSession())
}

// NewRGroups creates a new dai.Groups connected to the test database.
func NewRGroups() dai.Groups {
	return rethinkdb.NewRGroups(tdb.NewSession())
}

// NewRUsers creates a new dai.Users connected to the test database.
func NewRUsers() dai.Users {
	return rethinkdb.NewRUsers(tdb.NewSession())
}
