package ws

import (
	"github.com/emicklei/go-restful"
)

// NewRegisteredServicesContainer creates a container for all the web services.
func NewRegisteredServicesContainer() *restful.Container {
	wsContainer := restful.NewContainer()

	if err := newProjectResource(wsContainer); err != nil {
		panic("Could not register ProjectResource")
	}

	if err := newAdminResource(wsContainer); err != nil {
		panic("Could not register AdminResource")
	}

	return wsContainer
}
