package filter

import (
	"net/http"
	"sync"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/interfaces/db/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/doi"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/ws/rest"
)

type filterFunc func(request *restful.Request, response *restful.Response, chain *restful.FilterChain)

type apikeyCache struct {
	mutex         sync.RWMutex
	usersByAPIKey map[string]*schema.User
	apiKeyByUser  map[string]string
	users         doi.Users
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

func (c *apikeyCache) getUser(req *http.Request) (*rest.User, error) {
	u := rest.User{
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

func (c *apikeyCache) validUser(user *rest.User) bool {
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
