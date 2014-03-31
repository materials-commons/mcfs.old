package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request/handler"
)

func (h *ReqHandler) createProject(req *protocol.CreateProjectReq) (resp *protocol.CreateProjectResp, err error) {
	projHandler := handler.NewCreateProject(h.session)

	if !projHandler.Validate(req) {
		return nil, mc.Errorf(mc.ErrInvalid, "Invalid project name %s", req.Name)
	}

	proj, err := projHandler.GetProject(req.Name, h.user)
	switch {
	case err == nil:
		// Found project
		resp := &protocol.CreateProjectResp{
			ProjectID: proj.ID,
			DataDirID: proj.DataDir,
		}
		return resp, mc.ErrExists

	default:
		p, err := projHandler.CreateProject(req.Name, h.user)
		if err != nil {
			return nil, err
		}
		resp := &protocol.CreateProjectResp{
			ProjectID: p.ID,
			DataDirID: p.DataDir,
		}
		return resp, nil
	}
}
