package request

import (
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/server/protocol"
	"github.com/materials-commons/mcfs/server/service"
)

// login validates a login request.
func (h *ReqHandler) login(req *protocol.LoginReq) (*protocol.LoginResp, error) {
	if validLogin(req.User, req.APIKey) {
		h.user = req.User
		return &protocol.LoginResp{}, nil
	}

	return nil, mc.Errorf(mc.ErrInvalid, "Bad login %s/%s", req.User, req.APIKey)
}

// validLogin looks the user up in the database and compares the APIKey passed in with
// the APIKey in the database.
func validLogin(user, apikey string) bool {
	u, err := service.User.ByID(user)
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
func (h *ReqHandler) logout(req *protocol.LogoutReq) (*protocol.LogoutResp, error) {
	return &protocol.LogoutResp{}, nil
}
