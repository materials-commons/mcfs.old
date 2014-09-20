package ws

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/materialsd"
	"github.com/materials-commons/mcfs/materialsd/autoupdate"
	"github.com/materials-commons/mcfs/materialsd/config"
	"net/http"
	"os"
	"time"
)

type adminResource struct {
	updater *autoupdate.Updater
}

type updateStatus struct {
	Website bool `json:"website"`
	Server  bool `json:"server"`
}

func newAdminResource(container *restful.Container) error {
	adminResource := adminResource{
		updater: autoupdate.NewUpdater(),
	}
	adminResource.register(container)
	return nil
}

func (ar *adminResource) register(container *restful.Container) {
	ws := new(restful.WebService)
	ws.Path("/admin").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/restart").To(ar.restart).
		Doc("Restarts the materials service"))

	ws.Route(ws.GET("/update").To(ar.update).
		Doc("If updates are available downloads, installs and restarts the server."))

	ws.Route(ws.GET("/updates").Filter(JSONPFilter).To(ar.updates).
		Doc("List services that are available for update").
		Writes(updateStatus{}))

	ws.Route(ws.GET("/stop").To(ar.stop).
		Doc("Stops the server."))

	ws.Route(ws.GET("/config").Filter(JSONPFilter).To(ar.config).
		Doc("Returns the configuration settings").
		Writes(config.ConfigSettings{}))

	container.Add(ws)
}

func (ar *adminResource) restart(request *restful.Request, response *restful.Response) {
	response.WriteErrorString(http.StatusOK, "Restarting materials service\n")
	go func() {
		sleep(1)
		materials.Restart()
	}()
}

func (ar *adminResource) update(request *restful.Request, response *restful.Response) {
	websiteUpdate := "No"
	binaryUpdate := "No"

	if ar.updater.UpdatesAvailable() {
		if ar.updater.BinaryUpdate() {
			binaryUpdate = "Yes"
		}
	}

	msg := fmt.Sprintf("Website updated: %s/Binary updated: %s\n", websiteUpdate, binaryUpdate)
	response.WriteErrorString(http.StatusOK, msg)

	go func() {
		sleep(1)
		ar.updater.ApplyUpdates()
	}()
}

func (ar *adminResource) updates(request *restful.Request, response *restful.Response) {
	u := updateStatus{}
	if ar.updater.UpdatesAvailable() {
		if ar.updater.BinaryUpdate() {
			u.Server = true
		}
	}

	response.WriteEntity(u)
}

func (ar *adminResource) stop(request *restful.Request, response *restful.Response) {
	response.WriteErrorString(http.StatusOK, "Stopping materials service\n")
	go func() {
		sleep(1)
		os.Exit(0)
	}()
}

func (ar *adminResource) config(request *restful.Request, response *restful.Response) {
	response.WriteEntity(config.Config)
}

func sleep(seconds time.Duration) {
	time.Sleep(seconds * time.Millisecond)
}
