package filter

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/app"
)

// apikeyFilter holds the attributes of the apikey filter.
type apikeyFilter struct {
	users app.Users
}

// Filter implements the restful filter for apikeyFilter
func (f *apikeyFilter) Filter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	u, err := getUser(request.Request)
	if err != nil {
		response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
		return
	}

	if !f.users.IsValid(u) {
		response.WriteErrorString(http.StatusUnauthorized, "Not authorized")
		return
	}

	request.SetAttribute("user", u)
	chain.ProcessFilter(request, response)
}
