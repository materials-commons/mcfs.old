package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/request/handler"
	"strings"
)

func (h *ReqHandler) createDir(req *protocol.CreateDirReq) (resp *protocol.CreateResp, err error) {
	dh := handler.NewCreateDir(h.session)
	proj, err := dh.GetProject(req.ProjectID)
	switch {
	case err != nil:
		return nil, mc.Errorf(mc.ErrInvalid, "Bad projectID %s", req.ProjectID)
	case proj.Owner != h.user:
		return nil, mc.Errorf(mc.ErrNoAccess, "Access to project %s not allowed", req.ProjectID)
	case !validDirPath(proj.Name, req.Path):
		return nil, mc.Errorf(mc.ErrInvalid, "Invalid directory path %s", req.Path)
	default:
		dataDir, err := dh.GetDataDir(req)
		switch {
		case err == mc.ErrNotFound:
			var parent *schema.Directory
			if parent, err = dh.GetParent(req.Path); err != nil {
				return nil, mc.Errorm(mc.ErrNotFound, err)
			}
			dataDir, err := dh.CreateDir(req, h.user, parent.ID)
			if err != nil {
				return nil, mc.Errorm(mc.ErrInvalid, err)
			}
			resp := &protocol.CreateResp{
				ID: dataDir.ID,
			}
			return resp, nil
		case err != nil:
			return nil, mc.Errorm(mc.ErrNotFound, err)
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
