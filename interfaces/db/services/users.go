package services

import (
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/doi"
)

type usersService struct {
	users doi.Users
}

func NewUsersService(users doi.Users) usersService {
	return usersService{users: users}
}

func (s usersService) ByID(id string) (*schema.User, error) {
	return s.users.ByID(id)
}

func (s usersService) ByAPIKey(apikey string) (*schema.User, error) {
	return s.users.ByAPIKey(apikey)
}
