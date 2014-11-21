package rest

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/schema"
)

// httpError is the error and message to respond with.
type httpError struct {
	statusCode int
	message    string
}

// Write writes an httpError as the response.
func (e *httpError) Write(response *restful.Response) {
	response.WriteErrorString(e.statusCode, e.message)
}

// RouteFunc represents the routes function
type RouteFunc func(request *restful.Request, response *restful.Response, user schema.User) (error, interface{})

// Handler represents the way a route function should actually be written.
type Handler func(request *restful.Request, response *restful.Response)

// RouteHandler creates a wrapper function for route methods. This allows route
// methods to return errors and have them handled correctly.
func RouteHandler(f RouteFunc) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		user := schema.User{} //request.Attribute("user").(schema.User)
		err, val := f(request, response, user)
		switch {
		case err != nil:
			httpErr := errorToHTTPError(err)
			httpErr.Write(response)
		case val != nil:
			err = response.WriteEntity(val)
			if err != nil {
				// log the error here
			}
		default:
			// Nothing to do
		}
	}
}

// errorToHTTPError translates an error code into an httpError. It checks
// if the error code is of type mcerr.Error and handles it appropriately.
func errorToHTTPError(err error) *httpError {
	switch e := err.(type) {
	case *mcerr.Error:
		return appErrToHTTPError(e)
	default:
		return otherErrorToHTTPError(e)
	}
}

// mcerrorToHTTPError tranlates an mcerr.Error to an httpError.
func appErrToHTTPError(err *mcerr.Error) *httpError {
	httpErr := otherErrorToHTTPError(err.Err)
	httpErr.message = fmt.Sprintf("%s: %s", httpErr.message, err.Message)
	return httpErr
}

// otherErrorToHTTPError translates other error types to an httpError.
func otherErrorToHTTPError(err error) *httpError {
	var httpErr httpError
	switch err {
	case mcerr.ErrNotFound:
		httpErr.statusCode = http.StatusBadRequest
	case mcerr.ErrExists:
		httpErr.statusCode = http.StatusForbidden
	case mcerr.ErrNoAccess:
		httpErr.statusCode = http.StatusUnauthorized
	default:
		httpErr.statusCode = http.StatusInternalServerError
	}

	httpErr.message = err.Error()
	return &httpErr
}
