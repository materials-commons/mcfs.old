package request

import (
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/materials-commons/mcfs/protocol"
)

// statProject returns the list of entries (files and directories) for a given project.
// It can look up a project by its ID, or by its name. In the case of name lookup the
// owner of the project must match the user making the request.
func (h *ReqHandler) statProject(req *protocol.StatProjectReq) (*protocol.StatProjectResp, error) {
	var projectID string

	switch {
	case req.Name != "":
		// Lookup the project by its name.
		project, err := h.service.Project.ByName(req.Name, h.user)
		if err != nil {
			return nil, mcerr.Errorm(mcerr.ErrNotFound, err)
		}
		projectID = project.ID
	case req.ID != "":
		// Use the project id we were given.
		projectID = req.ID
	default:
		return nil, mcerr.Errorm(mcerr.ErrInvalid, nil)
	}

	entries, err := h.service.Project.Files(projectID, req.Base)
	if err != nil {
		return nil, mcerr.Errorm(mcerr.ErrNotFound, err)
	}

	resp := protocol.StatProjectResp{
		ProjectID: projectID,
		Entries:   entries,
	}
	return &resp, nil
}
