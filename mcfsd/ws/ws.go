package ws

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/ws/upload"
)

// NewRegisteredServicesContainer creates a container for all the web services.
func NewRegisteredServicesContainer() *restful.Container {
	container := restful.NewContainer()
	container.Add(upload.WebService())
	return container
}
