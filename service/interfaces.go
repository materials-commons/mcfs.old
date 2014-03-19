package service

import (
	"github.com/materials-commons/base/schema"
)

// ServiceDatabase specifies the type of database backend the service uses.
type ServiceDatabase int

const (
	// RethinkDB database backend
	RethinkDB ServiceDatabase = iota

	// SQL represents a generic SQL database backend
	SQL
)

// Users is the common API to users.
type Users interface {
	ByID(id string) (*schema.User, error)
	ByAPIKey(apikey string) (*schema.User, error)
	All() ([]schema.User, error)
}

// Files is the common API to files.
type Files interface {
	ByID(id string) (*schema.File, error)
	Update(*schema.File) error
	Insert(*schema.File) (*schema.File, error)
	AddDirectories(file *schema.File, dirIDs ...string) error
}

// Dirs is the common API to directories.
type Dirs interface {
	ByID(id string) (*schema.Directory, error)
	Update(*schema.Directory) error
	Insert(*schema.Directory) (*schema.Directory, error)
	AddFiles(dir *schema.Directory, fileIDs ...string) error
}
