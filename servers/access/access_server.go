package access

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs"
	"github.com/materials-commons/mcfs/service"
)

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

//
func Send(request *request) error {
	if !server.isRunning {
		return mcfs.ErrServerNotRunning
	}

	var err error = nil

	defer func() {
		if e := recover(); e != nil {
			err = mcfs.ErrServerNotRunning
		}
	}()

	server.request <- request
	return err
}

func Recv() *response {
	return <-server.response
}

// Init initializes the server. It meant to be called by the Server interface each
// time the server is started.
func (s *accessServer) Init() {
	s.request = make(chan *request)
	s.response = make(chan *response)
}

// Run implements the server. It is meant to be called by the Server interface.
func (s *accessServer) Run(stopChan <-chan struct{}) {
	s.isRunning = true
	if err := s.apikeys.load(); err != nil {
		s.shutdown()
		return
	}

	for {
		select {
		case request := <-s.request:
			s.doRequest(request)
		case <-stopChan:
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
			err:  mc.ErrNotFound,
		}
	}
}

// doInvalidRequest returns an error when the command is not recognized
func (s *accessServer) doInvalidRequest() {
	s.response <- &response{
		user: nil,
		err:  mc.ErrInvalid,
	}
}
