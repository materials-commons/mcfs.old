package ws

import (
	"net/http"
	"sync"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/service"
	"github.com/materials-commons/mcfs/schema"
)

type filterFunc func(request *restful.Request, response *restful.Response, chain *restful.FilterChain)

// User contains the user authentication information sent with the request.
type User struct {
	Name  string
	Token string
}

type apikeyCache struct {
	mutex         sync.RWMutex
	usersByAPIKey map[string]*schema.User
	apiKeyByUser  map[string]string
	users         service.Users
}

func (c *apikeyCache) apikeyFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	u, err := c.getUser(request.Request)
	if err != nil {
		response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
		return
	}

	if !c.validUser(u) {
		response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
		return
	}

	request.SetAttribute("user", *u)
	chain.ProcessFilter(request, response)
}

func (c *apikeyCache) getUser(req *http.Request) (*User, error) {
	u := User{
		Name:  getUsername(req),
		Token: getAPIKey(req),
	}

	switch {
	case u.Name == "":
		return nil, mcerr.ErrInvalid
	case u.Token == "":
		return nil, mcerr.ErrInvalid
	default:
		return &u, nil
	}
}

func (c *apikeyCache) validUser(user *User) bool {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	u := c.usersByAPIKey[user.Token]
	if u == nil {
		dbuser, err := c.users.ByAPIKey(user.Token)
		if err != nil || dbuser.Name != user.Name {
			return false
		}
		c.usersByAPIKey[user.Token] = dbuser
		c.apiKeyByUser[user.Name] = user.Token
	}
	return true
}
