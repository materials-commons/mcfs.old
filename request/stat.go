package request

import (
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
)

func (h *ReqHandler) stat(req *protocol.StatReq) (*protocol.StatResp, *stateStatus) {
	df, err := model.GetDataFile(req.DataFileID, h.session)
	switch {
	case err != nil:
		return nil, ssf(mc.ErrorCodeNotFound, "Unknown id %s", req.DataFileID)
	case !OwnerGaveAccessTo(df.Owner, h.user, h.session):
		return nil, ssf(mc.ErrorCodeNoAccess, "You do not have permission to access this datafile %s", req.DataFileID)
	default:
		return respStat(df), nil
	}
}

func respStat(df *schema.DataFile) *protocol.StatResp {
	return &protocol.StatResp{
		DataFileID: df.Id,
		Name:       df.Name,
		DataDirs:   df.DataDirs,
		Checksum:   df.Checksum,
		Size:       df.Size,
		Birthtime:  df.Birthtime,
		MTime:      df.MTime,
	}
}
