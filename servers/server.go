package servers

import (
	"launchpad.net/tomb"
)

// Instance is an instance of a server that is being monitored. A server implementing
// this interface must be restartable.
type Instance interface {
	// Run runs the server, passing in the channel it should listen on for
	// stop command.
	Run(stopChan <-chan struct{})

	// Init initializes server resources. In particular this will be called everytime
	// a server is started. It should re-initialize all server state.
	Init()
}

// Server is an instance of server that can be started, stopped and
// queried for status. A server is a go routine.
type Server struct {
	tomb.Tomb
	Instance
	started bool
}

// Start starts a server instance. It handles marking a server as done when it
// has finished running.
func (s *Server) Start() {
	s.Init()
	s.Tomb = tomb.Tomb{}
	s.started = true
	go func() {
		defer s.Done()
		s.Run(s.Dying())
	}()
}

// Stop stops a server instance.
func (s *Server) Stop() {
	s.started = false
	s.Kill(nil)
}

// Status returns the current status of the server.
func (s *Server) Status() Status {
	if !s.started {
		return Stopped
	}

	switch s.Err() {
	case tomb.ErrStillAlive:
		return Running
	case tomb.ErrDying:
		return Stopping
	default:
		return Stopped
	}
}
