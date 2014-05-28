package service

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/base/dir"
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/base/model"
	"github.com/materials-commons/mcfs/base/schema"
	"github.com/materials-commons/mcfs/server"
)

type rProjects struct{}

func newRProjects() rProjects {
	return rProjects{}
}

// ByID looks up a project by its primary key.
func (p rProjects) ByID(id string) (*schema.Project, error) {
	var project schema.Project
	if err := model.Projects.Q().ByID(id, &project); err != nil {
		return nil, mc.ErrNotFound
	}
	return &project, nil
}

// ByName looks up a project by its name and owner.
func (p rProjects) ByName(name, owner string) (*schema.Project, error) {
	var project schema.Project
	rql := model.Projects.T().GetAllByIndex("name", name).Filter(r.Row.Field("owner").Eq(owner))
	if err := model.Projects.Q().Row(rql, &project); err != nil {
		return nil, mc.ErrNotFound
	}
	return &project, nil
}

// Files returns a flattened list of all the files and directories in a project.
// Each entry has its full path starting from the project. The returned list is
// in sorted (ascending) order.
func (p rProjects) Files(projectID, base string) ([]dir.FileInfo, error) {
	rql := r.Table("project2datadir").GetAllByIndex("project_id", projectID).EqJoin("datadir_id", r.Table("datadirs_denorm")).Zip()
	var entries []schema.DataDirDenorm
	if err := model.DirsDenorm.Q().Rows(rql, &entries); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		// Nothing was found, treat as invalid project.
		return nil, mc.ErrNotFound
	}
	dirlist := &dirList{}
	return dirlist.build(entries, base), nil
}

// Update updates an existing project.
func (p rProjects) Update(project *schema.Project) error {
	return model.Projects.Q().Update(project.ID, project)
}

// Insert inserts a new project. This method creates the directory object
// for the project. If a directory id is specified in the project then
// the method will return ErrInvalid.
func (p rProjects) Insert(project *schema.Project) (*schema.Project, error) {
	if project.DataDir != "" {
		return nil, mc.ErrInvalid
	}

	var (
		newProject schema.Project
		newDir     *schema.Directory
		err        error
	)

	if err = model.Projects.Q().Insert(project, &newProject); err != nil {
		return nil, mcfs.ErrDBInsertFailed
	}

	dir := schema.NewDirectory(project.Name, project.Owner, newProject.ID, "")
	rdirs := newRDirs()

	if newDir, err = rdirs.Insert(&dir); err != nil {
		return nil, mcfs.ErrDBRelatedUpdateFailed
	}

	newProject.DataDir = newDir.ID
	if err = model.Projects.Q().Update(newProject.ID, &newProject); err != nil {
		return &newProject, err
	}

	err = p.AddDirectories(&newProject, newDir.ID)

	return &newProject, err
}

// AddDirectories adds new directories to the project.
func (p rProjects) AddDirectories(project *schema.Project, directoryIDs ...string) error {
	var rverror error
	// Add each directory to the project2datadir table. If there are any errors,
	// remember that we saw an error, but continue on.
	for _, dirID := range directoryIDs {
		p2d := schema.Project2DataDir{
			ProjectID: project.ID,
			DataDirID: dirID,
		}
		if err := model.Projects.Q().InsertRaw("project2datadir", p2d, nil); err != nil {
			rverror = mcfs.ErrDBRelatedUpdateFailed
		}
	}

	return rverror
}
