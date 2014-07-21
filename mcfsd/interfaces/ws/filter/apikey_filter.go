package filter

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/dai"
)

// apikeyFilter holds the attributes of the apikey filter.
type apikeyFilter struct {
	users dai.Users
}

// Filter implements the restful filter for apikeyFilter
func (f *apikeyFilter) Filter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	user, err := f.getValidUser(request)
	switch {
	case err != nil:
		response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
	default:
		request.SetAttribute("user", *user)
		chain.ProcessFilter(request, response)
	}
}

func (f *apikeyFilter) getValidUser(req *restful.Request) (*schema.User, error) {
	u, err := getUser(req.Request)
	if err != nil {
		return nil, err
	}

	user, err := f.users.ByAPIKey(u.apikey)
	if err != nil {
		return nil, err
	}

	if !user.IsValid(u.id, u.apikey) {
		return nil, mcerr.ErrNotAuthorized
	}

	return user, nil
}
