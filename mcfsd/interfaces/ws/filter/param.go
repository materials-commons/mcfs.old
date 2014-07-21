package filter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/materials-commons/mcfs/mcerr"
)

type user struct {
	id     string
	apikey string
}

// getUser retrieves the user information by retrieving the user and apikey
// params passed into the url point.
func getUser(req *http.Request) (user, error) {
	u := user{
		id:     getParam(req, "user"),
		apikey: getParam(req, "apikey"),
	}

	switch {
	case u.id == "":
		return u, mcerr.ErrInvalid
	case u.apikey == "":
		return u, mcerr.ErrInvalid
	default:
		return u, nil
	}
}

// getParam attempts to retrieve a param by first checking if it was passed
// in as a url param. If the value isn't found then it checks the header.
func getParam(req *http.Request, param string) string {
	value := req.FormValue(param)
	if value == "" {
		value = req.Header.Get(strings.ToUpper(fmt.Sprintf("MC-%s", param)))
	}

	return value
}
