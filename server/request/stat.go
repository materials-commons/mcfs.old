package request

import (
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/protocol"
)

// stat is like the file system stat call but returns information from our document store.
func (h *ReqHandler) stat(req *protocol.StatReq) (*protocol.StatResp, error) {
	file, err := h.dai.File.ByID(req.DataFileID)
	switch {
	case err != nil:
		return nil, mcerr.Errorf(mcerr.ErrNotFound, "Unknown id %s", req.DataFileID)
	case !h.dai.Group.HasAccess(file.Owner, h.user):
		return nil, mcerr.Errorf(mcerr.ErrNoAccess, "You do not have permission to access this datafile %s", req.DataFileID)
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
