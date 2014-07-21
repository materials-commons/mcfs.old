package ws

import (
	"github.com/emicklei/go-restful"
)

// NewRegisteredServicesContainer creates a container for all the web services.
func NewRegisteredServicesContainer() *restful.Container {
	container := restful.NewContainer()
	//	container.Add(upload.WebService())
	return container
}
