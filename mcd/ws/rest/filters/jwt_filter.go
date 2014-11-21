package filters

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

// jwtFilter implements a filter for JWT.
type jwtFilter struct {
	publicKey []byte // The key to use
	loginPath string // The login path to ignore (users acquire their token here)
}

// NewJWTFilter creates a new jwtFilter
func NewJWTFilter(publicKey []byte, loginPath string) *jwtFilter {
	return &jwtFilter{
		publicKey: publicKey,
		loginPath: loginPath,
	}
}

// Filter implements the logic of validating REST end points against the JWT token.
func (f *jwtFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// if the user is logging in for the first time then the
	// path will be f.loginPath. If that is the case then we just
	// go to the next filter because there is no token to
	// authenticate against.
	if req.Request.URL.Path != f.loginPath {
		token, err := jwt.ParseFromRequest(req.Request, f.getKey)
		if err != nil || !token.Valid {
			fmt.Printf("invalid token for url %s: %s\n ", req.Request.URL.Path, err)
			resp.WriteErrorString(http.StatusUnauthorized, "Not authorized")
			return
		}
	}
	chain.ProcessFilter(req, resp)
}

// Return the key jwt uses to validate a token.
func (f *jwtFilter) getKey(token *jwt.Token) (interface{}, error) {
	return f.publicKey, nil
}
