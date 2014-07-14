package service

import "github.com/emicklei/go-restful"

// REST implements a REST based web service.
type REST interface {
	WebService() *restful.WebService
}
