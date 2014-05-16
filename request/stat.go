package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"github.com/materials-commons/mcfs/service"
)

func (h *ReqHandler) stat(req *protocol.StatReq) (*protocol.StatResp, error) {
	df, err := model.GetFile(req.DataFileID, h.session)
	switch {
	case err != nil:
		return nil, mc.Errorf(mc.ErrNotFound, "Unknown id %s", req.DataFileID)
	case !service.Group.HasAccess(df.Owner, h.user):
		return nil, mc.Errorf(mc.ErrNoAccess, "You do not have permission to access this datafile %s", req.DataFileID)
	default:
		return respStat(df), nil
	}
}

func respStat(df *schema.File) *protocol.StatResp {
	return &protocol.StatResp{
		DataFileID: df.ID,
		Name:       df.Name,
		DataDirs:   df.DataDirs,
		Checksum:   df.Checksum,
		Size:       df.Size,
		Birthtime:  df.Birthtime,
		MTime:      df.MTime,
	}
}
