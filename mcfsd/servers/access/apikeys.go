package access

import (
	"github.com/materials-commons/mcfs/schema"
	"github.com/materials-commons/mcfs/server/service"
)

type apikeys struct {
	keys  map[string]schema.User
	users service.Users
}

func newAPIKeys(users service.Users) *apikeys {
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
