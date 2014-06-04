package access

import (
	"github.com/materials-commons/mcfs/base/log"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/server"
	"github.com/materials-commons/mcfs/server/service"
)

// Create our own context log that always includes our server name.
var l = log.New("server", "AccessServer")

// The command to send
type command int

// Command definitions
const (
	acGetUser command = iota
)

// Request to send
type request struct {
	command command
	arg     string
}

// Reponse to the request
type response struct {
	user *schema.User
	err  error
}

// accessServer defines a server to control access to a shared set of user keys.
// Communication and response are across channels. The server controls access
// to the shared set of keys. It allows us to control reloading and updating the keys
// in a safe manner.
type accessServer struct {
	isRunning bool
	apikeys   *apikeys
	request   chan *request
	response  chan *response
}

// We only expose a single access server. The public routines work against this instance.
var server = &accessServer{
	apikeys: newAPIKeys(service.NewUsers(service.RethinkDB)),
}

// Server returns the singleton accessServer.
func Server() *accessServer {
	return server
}

// Send sends a request to the server.
func (s *accessServer) Send(request *request) error {
	// Shortcut check, if we know the server isn't running then we
	// don't have to wait for the panic.
	if !s.isRunning {
		return mcfs.ErrServerNotRunning
	}

	var err error

	defer func() {
		if e := recover(); e != nil {
			l.Debug("Attempt to send when server is not running.")
			err = mcfs.ErrServerNotRunning
		}
	}()

	s.request <- request
	return err
}

// Recv receives a response from the server.
func (s *accessServer) Recv() *response {
	return <-s.response
}

// Init initializes the server. It meant to be called by the Server interface each
// time the server is started.
func (s *accessServer) Init() {
	s.request = make(chan *request)
	s.response = make(chan *response)
}

// Run implements the server. It is meant to be called by the Server interface.
func (s *accessServer) Run(stopChan <-chan struct{}) {
	l.Info("Starting")
	s.isRunning = true
	if err := s.apikeys.load(); err != nil {
		s.shutdown()
		l.Crit(log.Msg("Unable to load apikeys: %s\n", err))
		return
	}

	for {
		select {
		case request := <-s.request:
			l.Debug(log.Msg("Received request: %#v\n", request))
			s.doRequest(request)
		case <-stopChan:
			l.Info("Shutting down.")
			s.shutdown()
			return
		}
	}
}

// shutdown performs cleanup of the server data structures.
func (s *accessServer) shutdown() {
	s.isRunning = false
	close(s.request)
	close(s.response)
}

// doRequest performs the request sent along the channel. Unknown requests send back an error.
func (s *accessServer) doRequest(request *request) {
	switch request.command {
	case acGetUser:
		s.doGetUser(request.arg)
	default:
		l.Warn("Received invalid command:", "command", request.command)
		s.doInvalidRequest()
	}
}

// doGetUser looks up a user by their apikey
func (s *accessServer) doGetUser(apikey string) {
	u, found := s.apikeys.lookup(apikey)
	switch {
	case found:
		s.response <- &response{
			user: &u,
			err:  nil,
		}
	default:
		s.response <- &response{
			user: nil,
			err:  mcerr.ErrNotFound,
		}
	}
}

// doInvalidRequest returns an error when the command is not recognized
func (s *accessServer) doInvalidRequest() {
	s.response <- &response{
		user: nil,
		err:  mcerr.ErrInvalid,
	}
}
