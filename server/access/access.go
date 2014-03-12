package access

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/server"
	"github.com/materials-commons/mcfs/service"
	"launchpad.net/tomb"
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
	apikeys  *apikeys
	request  chan request
	response chan response
	tomb.Tomb
}

// We only expose a single access server. The public routines work against this instance.
var s = &accessServer{
	request:  make(chan request),
	response: make(chan response),
	apikeys:  newAPIKeys(service.NewUsers(service.RethinkDB)),
}

// Start starts the access server.
func Start() {
	s.Start()
}

// Stop stops the access server.
func Stop() {
	s.Stop()
}

// Status returns the servers current status
func Status() server.Status {
	return s.Status()
}

// GetUserByAPIKey returns the User for a given APIKey.
func GetUserByAPIKey(apikey string) (*schema.User, error) {
	request := request{
		command: acGetUser,
		arg:     apikey,
	}
	s.request <- request
	response := <-s.response
	return response.user, response.err
}

// server implements the access server.
func (s *accessServer) server() {
	defer s.Done()
	s.apikeys.load()

	for {
		select {
		case request := <-s.request:
			s.doRequest(&request)
		case <-s.Dying():
			close(s.request)
			close(s.response)
			return
		}
	}
}

// doRequest performs the request sent along the channel. Unknown requests send back an error.
func (s *accessServer) doRequest(request *request) {
	switch request.command {
	case acGetUser:
		u, found := s.apikeys.lookup(request.arg)
		if !found {
			s.response <- response{
				user: nil,
				err:  mc.ErrNotFound,
			}
		} else {
			s.response <- response{
				user: &u,
				err:  nil,
			}
		}
	default:
		s.response <- response{
			user: nil,
			err:  mc.ErrInvalid,
		}
	}
}

// Start starts the server.
func (s *accessServer) Start() {
	go s.server()
}

// Stop stops the server.
func (s *accessServer) Stop() {
	s.Kill(nil)
}

// Status returns the status of the server.
func (s *accessServer) Status() server.Status {
	switch s.Err() {
	case tomb.ErrStillAlive:
		return server.Running
	case tomb.ErrDying:
		return server.Stopping
	default:
		return server.Stopped
	}
}
