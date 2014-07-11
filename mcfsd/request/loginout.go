package request

import (
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/server/service"
)

// login validates a login request.
func (h *ReqHandler) login(req *protocol.LoginReq) (*protocol.LoginResp, error) {
	if validLogin(req.User, req.APIKey, h.service) {
		h.user = req.User
		return &protocol.LoginResp{}, nil
	}

	return nil, mcerr.Errorf(mcerr.ErrInvalid, "Bad login %s/%s", req.User, req.APIKey)
}

// validLogin looks the user up in the database and compares the APIKey passed in with
// the APIKey in the database.
func validLogin(user, apikey string, s *service.Service) bool {
	u, err := s.User.ByID(user)
	switch {
	case err != nil:
		return false
	case u.APIKey != apikey:
		return false
	default:
		return true
	}
}

// logout responds to a logout request. It currently doesn't do anything but the
// state machine will treat this request specially and will terminate.
func (h *ReqHandler) logout(req *protocol.LogoutReq) error {
	return nil
}
