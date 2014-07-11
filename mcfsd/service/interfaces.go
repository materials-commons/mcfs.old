package service

import (
	"github.com/materials-commons/mcfs/dir"
	"github.com/materials-commons/mcfs/schema"
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
	ByPath(name, dirID string) (*schema.File, error)
	ByPathChecksum(name, dirID, checksum string) ([]schema.File, error)
	ByPathPartials(name, dirID string) ([]schema.File, error)
	ByChecksum(checksum string) (*schema.File, error)
	MatchOn(key, value string) ([]schema.File, error)
	Hide(*schema.File) error
	Update(*schema.File) error
	Insert(file *schema.File) (*schema.File, error)
	InsertEntry(file *schema.File) (*schema.File, error)
	Delete(id string) error
	AddDirectories(file *schema.File, dirIDs ...string) error
}

// Dirs is the common API to directories.
type Dirs interface {
	ByID(id string) (*schema.Directory, error)
	ByPath(path, projectID string) (*schema.Directory, error)
	Update(*schema.Directory) error
	Insert(*schema.Directory) (*schema.Directory, error)
	AddFiles(dir *schema.Directory, fileIDs ...string) error
	RemoveFiles(dir *schema.Directory, fileIDs ...string) error
}

// Projects is the common API to projects.
type Projects interface {
	ByID(id string) (*schema.Project, error)
	ByName(name, owner string) (*schema.Project, error)
	Files(id, base string) ([]dir.FileInfo, error)
	Update(*schema.Project) error
	Insert(*schema.Project) (*schema.Project, error)
	AddDirectories(project *schema.Project, directoryIDs ...string) error
}

// Groups is the common API to groups.
type Groups interface {
	ByID(id string) (*schema.Group, error)
	Insert(*schema.Group) (*schema.Group, error)
	Delete(id string) error
	HasAccess(owner, user string) bool
}
