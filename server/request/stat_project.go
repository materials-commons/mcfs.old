package request

import (
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/server/protocol"
	"github.com/materials-commons/mcfs/server/service"
)

// statProject returns the list of entries (files and directories) for a given project.
// It can look up a project by its ID, or by its name. In the case of name lookup the
// owner of the project must match the user making the request.
func (h *ReqHandler) statProject(req *protocol.StatProjectReq) (*protocol.StatProjectResp, error) {
	var projectID string

	switch {
	case req.Name != "":
		// Lookup the project by its name.
		project, err := service.Project.ByName(req.Name, h.user)
		if err != nil {
			return nil, mc.Errorm(mc.ErrNotFound, err)
		}
		projectID = project.ID
	case req.ID != "":
		// Use the project id we were given.
		projectID = req.ID
	default:
		return nil, mc.Errorm(mc.ErrInvalid, nil)
	}

	entries, err := service.Project.Files(projectID, req.Base)
	if err != nil {
		return nil, mc.Errorm(mc.ErrNotFound, err)
	}

	resp := protocol.StatProjectResp{
		ProjectID: projectID,
		Entries:   entries,
	}
	return &resp, nil
}
