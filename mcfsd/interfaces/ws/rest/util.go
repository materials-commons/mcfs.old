package rest

import (
	"fmt"

	"github.com/emicklei/go-restful"
)

// HTTPError represents an error returned by a REST Service
type HTTPError struct {
	StatusCode int
	Message    string
}

// Write writes an HTTPError as the response.
func (e *HTTPError) Write(response *restful.Response) {
	response.WriteErrorString(e.StatusCode, e.Message)
}

// HTTPErrorm is a utility function for creating a new HTTPError.
func HTTPErrorm(status int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: status,
		Message:    message,
	}
}

// HTTPErrore is a utility function for creating a new HTTPError. It
// takes an error and turns it into a string.
func HTTPErrore(status int, err error) *HTTPError {
	return HTTPErrorm(status, err.Error())
}

// HTTPErrorf is a utility function for creating a new HTTPError. It constructs
// the error message from the set of args.
func HTTPErrorf(status int, message string, args ...interface{}) *HTTPError {
	msg := fmt.Sprintf(message, args...)
	return HTTPErrorm(status, msg)
}

// RouteFunc represents the routes function
type RouteFunc func(request *restful.Request, response *restful.Response) *HTTPError

// Handler represents the way a route function should actually be written.
type Handler func(request *restful.Request, response *restful.Response)

// RouteHandler creates a wrapper function for route methods. This allows route
// methods to return errors and have them handled correctly.
func RouteHandler(f RouteFunc) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		if err := f(request, response); err != nil {
			err.Write(response)
		}
	}
}
