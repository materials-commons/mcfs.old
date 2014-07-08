package create

import "github.com/emicklei/go-restful"

type createResource struct {
}

func NewResource(container *restful.Container) error {
	createResource := createResource{}
	createResource.register(container)
	return nil
}

func (r createResource) register(container *restful.Container) {
	ws := new(restful.WebService)

	ws.Path("/create").Produces(restful.MIME_JSON)
	// ws.Route(ws.GET("/file").To(r.createFile).
	// 	Consumes(protocol.File{}))
}
