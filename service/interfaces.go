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
	ByID(id string) (*schema.DataFile, error)
	Update(*schema.DataFile) error
	Insert(*schema.DataFile) (*schema.DataFile, error)
}

// Dirs is the common API to directories.
type Dirs interface {
	ByID(id string) (*schema.DataDir, error)
	Update(*schema.DataDir) error
	Insert(*schema.DataDir) (*schema.DataFile, error)
	AddFiles(dir *schema.DataDir, fileIds ...string) error
}
