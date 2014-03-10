package mcapi

import (
	"github.com/emicklei/go-restful"
	"net/http"

	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/base/model"
	"fmt"
)

type groupsResource struct {
}

func newGroupsResource(container *restful.Container) error {
	gr := groupsResource{}
	gr.register(container)
	return nil
}

func (g groupsResource) register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/groups").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").To(g.all).
		Doc("List all groups for user").
		Writes([]schema.UserGroup{}))

	container.Add(ws)
}

func (g groupsResource) all(request *restful.Request, response *restful.Response) {
	rql := model.Groups.T().GetAllByIndex("owner", "")
	var groups []schema.UserGroup
	if err := model.Groups.Q().Rows(rql, &groups); err != nil {
		response.WriteErrorString(http.StatusNotFound, fmt.Sprintf("Error querying database %s", err))
		return
	}

	response.WriteEntity(groups)
}
