package server

import (
	"launchpad.net/tomb"
)

// Instance is an instance of a server that is being monitored.
type Instance interface {
	Run(stopChan <-chan struct{})
}

// Server is an instance of server that can be started, stopped and
// queried for status. A server is a go routine.
type Server struct {
	tomb.Tomb
	Instance
}

// Start starts a server instance. It handles marking a server as done when it
// has finished running.
func (s *Server) Start() {
	go func() {
		defer s.Done()
		s.Run(s.Dying())
	}()
}

func (s *Server) Stop() {
	s.Kill(nil)
}

func (s *Server) Status() Status {
	switch s.Err() {
	case tomb.ErrStillAlive:
		return Running
	case tomb.ErrDying:
		return Stopping
	default:
		return Stopped
	}
}
