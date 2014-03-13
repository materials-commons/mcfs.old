package recover

import ()

type Request struct{}

type Response struct{}

type recoveryServer struct {
	request chan Request
}

var server = &recoveryServer{
	request: make(chan Request, 50),
}

func Server() *recoveryServer {
	return server
}

func (s *recoveryServer) Run(stopChan <-chan struct{}) {

}
