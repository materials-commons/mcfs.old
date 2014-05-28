package ws

/*
 * monitor.go contains the routines for monitoring project files and directory changes. It
 * communicates with the frontend using socket.io. Each project is monitored separately.
 */

import (
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/materials-commons/gohandy/fs"
	"github.com/materials-commons/mcfs/client"
	"github.com/materials-commons/mcfs/client/config"
	"net/http"
	"os"
	"time"
)

var _ = fmt.Println

// projectFileStatus communicates the types of changes that
// have occured.
type projectFileStatus struct {
	Project  string `json:"project"`
	FilePath string `json:"filepath"`
	Event    string `json:"event"`
}

// startMonitor starts the monitor service and the HTTP and SocketIO connections.
func startMonitor() {
	sio := socketio.NewSocketIOServer(&socketio.Config{})
	sio.On("connect", func(ns *socketio.NameSpace) {
		// Nothing to do
	})
	sio.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Nothing to do
	})
	go monitorProjectChanges(sio)
	go startHTTP(10, sio)
}

// startHttp starts up a HTTP server. It will attempt to start the server
// retryCount times. The retry on server startup handles the case where
// the old materials service is stopping, and the new one is starting.
func startHTTP(retryCount int, sio *socketio.SocketIOServer) {
	for i := 0; i < retryCount; i++ {
		address := fmt.Sprintf(":%d", config.Config.Server.SocketIOPort)
		fmt.Println(http.ListenAndServe(address, sio))
		time.Sleep(1000 * time.Millisecond)
	}
	os.Exit(1)
}

// monitorProjectChanges starts a separate go thread for each project to monitor.
func monitorProjectChanges(sio *socketio.SocketIOServer) {
	p := materials.CurrentUserProjectDB()

	for _, project := range p.Projects() {
		go projectWatcher(project, p, sio)
	}
}

// projectWatcher starts the file system monitor. It watches for file system
// events and then communicates them along the SocketIOServer. It sends events
// to the front end as projectFileStatus messages encoded in JSON.
func projectWatcher(project *materials.Project, projectdb *materials.ProjectDB, sio *socketio.SocketIOServer) {
	for {
		watcher, err := fs.NewRecursiveWatcher(project.Path)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}
		watcher.Start()

	FsEventsLoop:
		for {
			select {
			case event := <-watcher.Events:
				broadcastEvent(event, project.Name, sio)
				trackEvent(event, project, projectdb)
			case err := <-watcher.ErrorEvents:
				fmt.Println("error:", err)
				break FsEventsLoop
			}
		}
		watcher.Close()
	}
}

// broadcastEvent takes a file system event and broadcasts it on socketIO.
func broadcastEvent(event fs.Event, projectName string, sio *socketio.SocketIOServer) {
	eventType := eventType2String(event)
	pfs := &projectFileStatus{
		Project:  projectName,
		FilePath: event.Name,
		Event:    eventType,
	}
	sio.Broadcast("file", pfs)
}

// trackEvent persists a file system event for a project to the projects list of events file.
func trackEvent(event fs.Event, project *materials.Project, projectdb *materials.ProjectDB) {
	if event.IsCreate() || event.IsModify() || event.IsDelete() {
		eventType := eventType2String(event)
		fileChange := materials.ProjectFileChange{
			Path: event.Name,
			Type: eventType,
			When: time.Now(),
		}
		projectdb.Update(func() *materials.Project {
			project.AddFileChange(fileChange)
			return project
		})
	}
}

// eventType takes an event determines what type of event occurred and
// returns the corresponding string.
func eventType2String(event fs.Event) string {
	switch {
	case event.IsCreate():
		return "Created"
	case event.IsDelete():
		return "Deleted"
	case event.IsModify():
		return "Modified"
	case event.IsRename():
		return "Renamed"
	default:
		return "Unknown"
	}
}
