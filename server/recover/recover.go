package recover

import (
	"launchpad.net/tomb"
)

type Request struct{}

type Response struct{}

type recoverServer struct {
	request chan Request
	tomb.Tomb
}

var s = &recoverServer{
	request: make(chan Request, 50),
}

// Start starts the recover server.
func Start() {
	//s.Start()
}

// Stop stops the recover server.
func Stop() {
	//s.Stop()
}
