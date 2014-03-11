package service

import (
	"github.com/materials-commons/base/schema"
)

type ServiceDatabase int

const (
	RethinkDB ServiceDatabase = iota
	SQL
)

type Users interface {
	ByID(id string) (*schema.User, error)
	ByAPIKey(apikey string) (*schema.User, error)
	All() (*[]schema.User, error)
}
