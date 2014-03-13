package recover

import ()

type Request struct{}

type Response struct{}

type recoverServer struct {
	request chan Request
}

var server = &recoverServer{
	request: make(chan Request, 50),
}

func Server() *recoverServer {
	return server
}

func (s *recoverServer) Run(stopChan <-chan struct{}) {

}
