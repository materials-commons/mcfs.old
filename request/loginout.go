package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/mcfs/protocol"
)

func (h *ReqHandler) login(req *protocol.LoginReq) (*protocol.LoginResp, error) {
	if validLogin(req.User, req.APIKey, h.session) {
		h.user = req.User
		return &protocol.LoginResp{}, nil
	}

	return nil, mc.Errorf(mc.ErrInvalid, "Bad login %s/%s", req.User, req.APIKey)
}

func validLogin(user, apikey string, session *r.Session) bool {
	u, err := model.GetUser(user, session)
	switch {
	case err != nil:
		return false
	case u.APIKey != apikey:
		return false
	default:
		return true
	}
}

func (h *ReqHandler) logout(req *protocol.LogoutReq) (*protocol.LogoutResp, error) {
	return &protocol.LogoutResp{}, nil
}
