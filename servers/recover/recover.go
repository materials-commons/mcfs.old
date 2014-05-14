package recover

import ()

// Request is what we have been asked to do.
type Request struct{}

// Response is something else
type Response struct{}

type recoveryServer struct {
	request chan Request
}

var server = &recoveryServer{
	request: make(chan Request, 50),
}

// Server returns the server
func Server() *recoveryServer {
	return server
}

// Run runs the server
func (s *recoveryServer) Run(stopChan <-chan struct{}) {

}
