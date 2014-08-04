package services

import (
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
)

type usersService struct {
	users dai.Users
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
