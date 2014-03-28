package service

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/dir"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs"
)

type rProjects struct{}

func newRProjects() rProjects {
	return rProjects{}
}

// ByID looks up a project by its primary key.
func (p rProjects) ByID(id string) (*schema.Project, error) {
	var project schema.Project
	if err := model.Projects.Q().ByID(id, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// Files returns a flattened of all the files and directories in a project.
// Each entry has its full path starting from the project. The returned list
// is in sorted (ascending) order.
func (p rProjects) Files(projectID string) ([]*dir.FileInfo, error) {
	rql := r.Table("project2datadir").GetAllByIndex("project_id", projectID).EqJoin("datadir_id", r.Table("datadirs_denorm")).Zip()
	var entries []schema.DataDirDenorm
	if err := model.DirsDenorm.Q().Rows(rql, &entries); err != nil {
		return nil, err
	}
	dlist := &dirList{}
	return dlist.build(entries), nil
}

// Update updates an existing project.
func (p rProjects) Update(project *schema.Project) error {
	return model.Projects.Q().Update(project.ID, project)
}

// Insert inserts a new project.
func (p rProjects) Insert(project *schema.Project) (*schema.Project, error) {
	var newProject schema.Project
	if err := model.Projects.Q().Insert(project, &newProject); err != nil {
		return nil, mcfs.ErrDBInsertFailed
	}
	return &newProject, nil
}
