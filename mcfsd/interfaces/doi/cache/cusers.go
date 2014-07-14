package cache

import (
	"sync"

	"github.com/materials-commons/mcfs/interfaces/db/schema"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/doi"
)

// userLookupFunc specifies the function to use to look up a user
type userLookupFunc func(id string) (*schema.User, error)

// cUsers implements the Users interface. It caches users in a set of maps
// for fast look ups. All methods on this type are thread safe.
type cUsers struct {
	mutex         sync.RWMutex            // Mutex to provide safe access.
	usersByAPIKey map[string]*schema.User // Lookup users by their API Key.
	usersByID     map[string]*schema.User // Lookup users by their id
	dbusers       doi.Users               // Interface to retrieve users from the database.
}

// NewCUsers creates a new cached user. The dbusers parameter is the interface
// to use to lookup users in the database.
func NewCUsers(dbusers doi.Users) *cUsers {
	return &cUsers{
		dbusers: dbusers,
	}
}

// ByID looks up a user in the cache by their id. If the user isn't in the cache then
// it looks them up using the dbusers interface.
func (u *cUsers) ByID(id string) (*schema.User, error) {
	user := u.lookupByID(id)
	if user != nil {
		return user.Clone(), nil
	}

	return u.dblookupByFunc(id, u.dbusers.ByID)
}

// lookupByID performs the look up of the user in the usersByID hash map.
func (u *cUsers) lookupByID(id string) *schema.User {
	defer u.mutex.RUnlock()
	u.mutex.RLock()
	return u.usersByID[id]
}

// dblookupByFunc looks a user up by the provided function. If the user is found
// then they are inserted into the set of caches.
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

// ByAPIKey looks up a user in the cache by their id. If the user isn't in the cache then
// it looks them up using the dbusers interface.
func (u *cUsers) ByAPIKey(apikey string) (*schema.User, error) {
	user := u.lookupByAPIKey(apikey)
	if user != nil {
		return user.Clone(), nil
	}

	return u.dblookupByFunc(apikey, u.dbusers.ByAPIKey)
}

// lookupByAPIKey performs the look up of the user in the usersByAPIKey hash map.
func (u *cUsers) lookupByAPIKey(apikey string) *schema.User {
	defer u.mutex.RUnlock()
	u.mutex.RLock()
	return u.usersByAPIKey[apikey]
}

// All returns all users in the cache.
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
