package filter

import (
	"fmt"
	"net/http"
	"strings"
)

func getAPIKey(req *http.Request) string {
	return getParam(req, "apikey")
}

func getUsername(req *http.Request) string {
	return getParam(req, "user")
}

func getParam(req *http.Request, param string) string {
	value := req.FormValue(param)
	if value == "" {
		value = req.Header.Get(strings.ToUpper(fmt.Sprintf("MC-%s", param)))
	}

	return value
}
