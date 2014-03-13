package access

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
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
	apikeys  *apikeys
	request  chan request
	response chan response
}

// We only expose a single access server. The public routines work against this instance.
var server = &accessServer{
	request:  make(chan request),
	response: make(chan response),
	apikeys:  newAPIKeys(service.NewUsers(service.RethinkDB)),
}

func Server() *accessServer {
	return server
}

// server implements the access server.
func (s *accessServer) Run(stopChan <-chan struct{}) {
	s.apikeys.load()
	for {
		select {
		case request := <-s.request:
			s.doRequest(&request)
		case <-stopChan:
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
		s.response <- response{
			user: &u,
			err:  nil,
		}
	default:
		s.response <- response{
			user: nil,
			err:  mc.ErrNotFound,
		}
	}
}

// doInvalidRequest returns an error when the command is not recognized
func (s *accessServer) doInvalidRequest() {
	s.response <- response{
		user: nil,
		err:  mc.ErrInvalid,
	}
}
