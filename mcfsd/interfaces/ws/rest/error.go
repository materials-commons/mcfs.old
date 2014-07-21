package rest

import "github.com/emicklei/go-restful"

type RESTError struct {
	Message    string
	HTTPStatus int
}

func (e *RESTError) WriteError(response *restful.Response) {
	response.WriteErrorString(e.HTTPStatus, e.Message)
}
