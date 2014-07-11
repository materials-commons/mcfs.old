package mcfs

import (
	"github.com/materials-commons/mcfs/dir"
	"github.com/materials-commons/mcfs/protocol"
)

// ProjectStat describes the project on the server.
type ProjectStat struct {
	ID      string         // ID of project on server
	Name    string         // Name of project
	Entries []dir.FileInfo // All files and directories in project
}

// StatProject sends a request to the server to get its view of the project.
func (c *Client) StatProject(projectName string) (*ProjectStat, error) {
	req := protocol.StatProjectReq{
		Name: projectName,
		Base: "/home/gtarcea",
	}

	resp, err := c.doRequest(req)
	if resp == nil {
		return nil, err
	}

	switch t := resp.(type) {
	case protocol.StatProjectResp:
		return &ProjectStat{
			Name:    projectName,
			ID:      t.ProjectID,
			Entries: t.Entries,
		}, nil
	default:
		return nil, ErrBadResponseType
	}
}
