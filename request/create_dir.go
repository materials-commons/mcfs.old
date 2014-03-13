package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request/handler"
	"strings"
)

func (h *ReqHandler) createDir(req *protocol.CreateDirReq) (resp *protocol.CreateResp, s *stateStatus) {
	dh := handler.NewCreateDir(h.session)
	proj, err := dh.GetProject(req.ProjectID)
	switch {
	case err != nil:
		return nil, ssf(mc.ErrorCodeInvalid, "Bad projectID %s", req.ProjectID)
	case proj.Owner != h.user:
		return nil, ssf(mc.ErrorCodeNoAccess, "Access to project not allowed")
	case !validDirPath(proj.Name, req.Path):
		return nil, ssf(mc.ErrorCodeInvalid, "Invalid directory path %s", req.Path)
	default:
		dataDir, err := dh.GetDataDir(req)
		switch {
		case err == mc.ErrNotFound:
			var parent *schema.DataDir
			if parent, err = dh.GetParent(req.Path); err != nil {
				return nil, ss(mc.ErrorCodeNotFound, err)
			}
			dataDir, err := dh.CreateDir(req, h.user, parent.ID)
			if err != nil {
				return nil, ss(mc.ErrorCodeInvalid, err)
			}
			resp := &protocol.CreateResp{
				ID: dataDir.ID,
			}
			return resp, nil
		case err != nil:
			return nil, ss(mc.ErrorCodeNotFound, err)
		default:
			resp := &protocol.CreateResp{
				ID: dataDir.ID,
			}
			return resp, nil
		}
	}
}

func validDirPath(projName, dirPath string) bool {
	slash := strings.Index(dirPath, "/")
	if slash == -1 {
		slash = strings.Index(dirPath, "\\")
	}
	switch {
	case slash == -1:
		return false
	case projName != dirPath[:slash]:
		return false
	default:
		return true
	}
}
