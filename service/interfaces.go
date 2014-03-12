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

// Users is the common API that all database implementations provide
type Users interface {
	ByID(id string) (*schema.User, error)
	ByAPIKey(apikey string) (*schema.User, error)
	All() ([]schema.User, error)
}
