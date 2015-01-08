package app

import "github.com/materials-commons/mcfs/base/schema"

// UsersService encapsulates user retrieval.
type UsersService interface {
	ByID(id string) (*schema.User, error)
	ByAPIKey(apikey string) (*schema.User, error)
}
