package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request/handler"
)

func (h *ReqHandler) createProject(req *protocol.CreateProjectReq) (resp *protocol.CreateProjectResp, s *stateStatus) {
	projHandler := handler.NewCreateProject(h.session)

	if !projHandler.Validate(req) {
		s = ssf(mc.ErrorCodeInvalid, "Invalid project name %s", req.Name)
		return nil, s
	}

	proj, err := projHandler.GetProject(req.Name, h.user)
	switch {
	case err == nil:
		// Found project
		resp := &protocol.CreateProjectResp{
			ProjectID: proj.ID,
			DataDirID: proj.DataDir,
		}
		return resp, ss(mc.ErrorCodeExists, mc.ErrExists)

	default:
		p, err := projHandler.CreateProject(req.Name, h.user)
		if err != nil {
			s.status = mc.ErrorCodeCreate
			s.err = err
			return nil, s
		}
		resp := &protocol.CreateProjectResp{
			ProjectID: p.ID,
			DataDirID: p.DataDir,
		}
		return resp, nil
	}
}
