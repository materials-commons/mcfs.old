package mcfs

import (
	"fmt"
	"github.com/materials-commons/mcfs/protocol"
)

func (c *Client) CreateDir(projectID, path string) (dataDirID string, err error) {
	req := &protocol.CreateDirReq{
		ProjectID: projectID,
		Path:      path,
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	switch t := resp.(type) {
	case protocol.CreateResp:
		return t.ID, nil
	default:
		fmt.Printf("2 %s %T\n", ErrBadResponseType, t)
		return "", ErrBadResponseType
	}
}

func (c *Client) UploadDirectory(projectID, dataDirID string, path string) ([]DataFileUpload, error) {
	switch {
	case dataDirID == "":
		return c.uploadNewDirectory(projectID, path)
	default:
		return c.uploadExistingDirectory(projectID, dataDirID, path)
	}
}

func (c *Client) uploadNewDirectory(projectID, path string) ([]DataFileUpload, error) {
	return nil, nil
}

func (c *Client) uploadExistingDirectory(projectID, dataDirID, path string) ([]DataFileUpload, error) {

	return nil, nil
}
