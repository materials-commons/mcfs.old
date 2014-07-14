package sync

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/protocol"
)

type syncResource struct {
}

// WebService creates an instance of the sync webservice.
func WebService() *restful.WebService {
	r := syncResource{}
	ws := new(restful.WebService)

	ws.Path("/sync").Produces(restful.MIME_JSON)
	ws.Route(ws.PUT("/start/{project-id}").To(r.syncStart).
		Doc("Attempts to acquire a sync token").
		Param(ws.PathParameter("project-id", "id of project to sync").DataType("string")).
		Writes(protocol.SyncStartResp{}))
	ws.Route(ws.PUT("/done/{sync-token-id}").To(r.syncDone).
		Doc("Finishes a sync request and releases the sync token").
		Param(ws.PathParameter("sync-token-id", "sync token recieved from the sync/start call")))
	ws.Route(ws.GET("/status/{project-id}").To(r.syncProjectStatus).
		Doc("Get the sync status of a project").
		Param(ws.PathParameter("project-id", "id of project to check")).
		Writes(protocol.SyncStatusResp{}))

	return ws
}

func (r syncResource) syncStart(request *restful.Request, response *restful.Response) {

}

func (r syncResource) syncDone(request *restful.Request, response *restful.Response) {

}

func (r syncResource) syncProjectStatus(request *restful.Request, response *restful.Response) {

}
