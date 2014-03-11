package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
)

type modelValidator struct {
	session *r.Session
	user    string
}

func newModelValidator(user string, session *r.Session) modelValidator {
	return modelValidator{
		session: session,
		user:    user,
	}
}

func (v modelValidator) datadirInProject(datadirID, projectID string) bool {
	query := r.Table("project2datadir").GetAllByIndex("datadir_id", datadirID)
	var p2d schema.Project2DataDir
	err := model.GetRow(query, v.session, &p2d)
	switch {
	case err != nil:
		return false
	case p2d.ProjectID != projectID:
		return false
	default:
		return true
	}
}

func (v modelValidator) datafileExistsInDataDir(datadirID, datafileName string) bool {
	rows, err := r.Table("datafiles").GetAllByIndex("name", datafileName).Run(v.session)
	if err != nil {
		return true // don't know if it exists or not
	}
	defer rows.Close()

	for rows.Next() {
		var df schema.DataFile
		rows.Scan(&df)
		for _, ddirID := range df.DataDirs {
			if datadirID == ddirID {
				return true
			}
		}
	}
	return false
}

func (v modelValidator) verifyProject(projectID string) bool {
	project, err := model.GetProject(projectID, v.session)
	switch {
	case err != nil:
		return false
	case project.Owner != v.user:
		return false
	default:
		return true
	}
}
