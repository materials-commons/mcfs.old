package cache

import (
	"sync"

	"github.com/materials-commons/mcfs/interfaces/db/schema"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/doi"
)

type userLookupFunc func(id string) (*schema.User, error)

type cUsers struct {
	mutex         sync.RWMutex
	usersByAPIKey map[string]*schema.User
	usersByID     map[string]*schema.User
	dbusers       doi.Users
}

func NewCUsers(dbusers doi.Users) *cUsers {
	return &cUsers{
		dbusers: dbusers,
	}
}

func (u *cUsers) ByID(id string) (*schema.User, error) {
	user := u.lookupByID(id)
	if user != nil {
		return user.Clone(), nil
	}

	return u.dblookupByFunc(id, u.dbusers.ByID)
}

func (u *cUsers) lookupByID(id string) *schema.User {
	defer u.mutex.RUnlock()
	u.mutex.RLock()
	return u.usersByID[id]
}

func (u *cUsers) dblookupByFunc(id string, flookup userLookupFunc) (*schema.User, error) {
	user, err := flookup(id)
	if err != nil {
		return nil, err
	}

	defer u.mutex.Unlock()
	u.mutex.Lock()
	u.usersByID[id] = user
	u.usersByAPIKey[user.APIKey] = user
	return user.Clone(), nil
}

func (u *cUsers) ByAPIKey(apikey string) (*schema.User, error) {
	user := u.lookupByAPIKey(apikey)
	if user != nil {
		return user.Clone(), nil
	}

	return u.dblookupByFunc(apikey, u.dbusers.ByAPIKey)
}

func (u *cUsers) lookupByAPIKey(apikey string) *schema.User {
	defer u.mutex.RUnlock()
	u.mutex.RLock()
	return u.usersByAPIKey[apikey]
}

func (u *cUsers) All() ([]schema.User, error) {
	defer u.mutex.RUnlock()
	u.mutex.RLock()

	var users []schema.User
	for _, user := range u.usersByID {
		cuser := user.Clone()
		users = append(users, *cuser)
	}
	return users, nil
}
