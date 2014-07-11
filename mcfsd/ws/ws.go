package ws

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/mcfsd/ws/upload"
)

// NewRegisteredServicesContainer creates a container for all the web services.
func NewRegisteredServicesContainer() *restful.Container {
	container := restful.NewContainer()

	if err := upload.NewResource(container); err != nil {
		panic("Could not register upload resource")
	}

	return container
}
