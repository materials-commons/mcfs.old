package files

import (
	"github.com/emicklei/go-restful"
	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/mcfsd/app"
	"github.com/materials-commons/mcfs/protocol"
)

// createFile creates a new file, or returns an existing file.
func (r *filesResource) createFile(request *restful.Request, response *restful.Response, user schema.User) (error, interface{}) {
	var req protocol.CreateFileReq

	if err := request.ReadEntity(&req); err != nil {
		return err, nil
	}

	f := app.File{
		Name:        req.Name,
		ProjectID:   req.ProjectID,
		DirectoryID: req.DirectoryID,
		Checksum:    req.Checksum,
		Size:        req.Size,
		Owner:       user.Name,
	}

	file, err := r.files.Create(f)
	if err != nil && err != mcerr.ErrExists {
		return err, nil
	}

	return nil, file
}
