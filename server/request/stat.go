package request

import (
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/server/protocol"
	"github.com/materials-commons/mcfs/server/service"
)

// stat is like the file system stat call but returns information from our document store.
func (h *ReqHandler) stat(req *protocol.StatReq) (*protocol.StatResp, error) {
	file, err := service.File.ByID(req.DataFileID)
	switch {
	case err != nil:
		return nil, mc.Errorf(mc.ErrNotFound, "Unknown id %s", req.DataFileID)
	case !service.Group.HasAccess(file.Owner, h.user):
		return nil, mc.Errorf(mc.ErrNoAccess, "You do not have permission to access this datafile %s", req.DataFileID)
	default:
		return respStat(file), nil
	}
}

// respStat creates the StatResp object from the file.
func respStat(file *schema.File) *protocol.StatResp {
	return &protocol.StatResp{
		DataFileID: file.ID,
		Name:       file.Name,
		DataDirs:   file.DataDirs,
		Checksum:   file.Checksum,
		Size:       file.Size,
		Birthtime:  file.Birthtime,
		MTime:      file.MTime,
	}
}
