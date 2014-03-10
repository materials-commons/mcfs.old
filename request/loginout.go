package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/mcfs/protocol"
)

func (h *ReqHandler) login(req *protocol.LoginReq) (*protocol.LoginResp, *stateStatus) {
	if validLogin(req.User, req.ApiKey, h.session) {
		h.user = req.User
		return &protocol.LoginResp{}, nil
	} else {
		return nil, ssf(mc.ErrorCodeInvalid, "Bad login %s/%s", req.User, req.ApiKey)
	}
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

func (r *ReqHandler) logout(req *protocol.LogoutReq) (*protocol.LogoutResp, *stateStatus) {
	return &protocol.LogoutResp{}, nil
}
