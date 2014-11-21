package access

import (
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/server/dai"
)

type apikeys struct {
	keys  map[string]schema.User
	users dai.Users
}

func newAPIKeys(users dai.Users) *apikeys {
	return &apikeys{
		keys:  make(map[string]schema.User),
		users: users,
	}
}

func (a *apikeys) load() error {
	users, err := a.users.All()
	if err != nil {
		return err
	}

	for _, user := range users {
		a.keys[user.APIKey] = user
	}

	return nil
}

func (a *apikeys) reload() error {
	a.keys = make(map[string]schema.User)
	return a.load()
}

func (a *apikeys) lookup(apikey string) (user schema.User, found bool) {
	user, found = a.keys[apikey]
	return user, found
}
