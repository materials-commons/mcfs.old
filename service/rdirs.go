package service

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/gohandy/arrays"
	"github.com/materials-commons/mcfs"
)

// rDirs implements the Dirs interface for RethinkDB
type rDirs struct{}

// newRDirs creates a new instance of rDirs
func newRDirs() rDirs {
	return rDirs{}
}

// ByID looks up a dir by its primary key. In RethinkDB this is the id field.
func (d rDirs) ByID(id string) (*schema.Directory, error) {
	var dir schema.Directory
	if err := model.Dirs.Q().ByID(id, &dir); err != nil {
		return nil, err
	}
	return &dir, nil
}

// ByPath looks up a directory in a project by its path.
func (d rDirs) ByPath(path, projectID string) (*schema.Directory, error) {
	rql := model.Dirs.T().GetAllByIndex("name", path).Filter(r.Row.Field("project").Eq(projectID))
	var dir schema.Directory
	if err := model.Dirs.Q().Row(rql, &dir); err != nil {
		return nil, err
	}
	return &dir, nil
}

// Update updates an existing dir. If you are adding new files you should use the
// AddFiles method. This method will not update related items. AddFiles takes care
// of updating other related tables.
func (d rDirs) Update(dir *schema.Directory) error {
	if err := model.Dirs.Q().Update(dir.ID, dir); err != nil {
		return err
	}
	return nil
}

// Insert creates a new dir. This method can return an error, with a valid
// DataDir. This happens when the dir is created, but one or more of the intermediate
// steps failed. The call needs to handle this case.
func (d rDirs) Insert(dir *schema.Directory) (*schema.Directory, error) {
	var newDir schema.Directory
	if err := model.Dirs.Q().Insert(dir, &newDir); err != nil {
		return nil, mcfs.ErrDBInsertFailed
	}

	// Insert the directory into the denorm table.
	var ddirDenorm = schema.DataDirDenorm{
		ID:        newDir.ID,
		Name:      newDir.Name,
		Owner:     newDir.Owner,
		Birthtime: newDir.Birthtime,
	}

	if len(newDir.DataFiles) > 0 {
		var err error
		if ddirDenorm.DataFiles, err = d.createDataFiles(newDir.DataFiles); err != nil {
			return &newDir, mcfs.ErrDBRelatedUpdateFailed
		}
	}

	if err := model.DirsDenorm.Q().Insert(ddirDenorm, nil); err != nil {
		return &newDir, mcfs.ErrDBRelatedUpdateFailed
	}

	return &newDir, nil
}

// AddFiles adds new file ids to a dir. It updates all related items and join tables.
// AddFiles will return ErrRelatedUpdateFailed or ErrUpdateFailed when an error occurs.
// The caller will have to decide how to handle these errors because the database will
// be out of sync.
func (d rDirs) AddFiles(dir *schema.Directory, fileIDs ...string) error {
	// Add fileIds to the Directory
	for _, id := range fileIDs {
		dir.DataFiles = append(dir.DataFiles, id)
	}

	if err := model.Dirs.Q().Update(dir.ID, dir); err != nil {
		return mcfs.ErrDBUpdateFailed
	}

	// Add entries to the denorm table for this dir.
	var dirDenorm schema.DataDirDenorm
	newEntries, err := d.createDataFiles(fileIDs)
	if err != nil {
		return mcfs.ErrDBRelatedUpdateFailed
	}

	if err := model.DirsDenorm.Q().ByID(dir.ID, &dirDenorm); err != nil {
		return mcfs.ErrDBRelatedUpdateFailed
	}

	dirDenorm.DataFiles = append(dirDenorm.DataFiles, newEntries...)
	if err := model.DirsDenorm.Q().Update(dirDenorm.ID, dirDenorm); err != nil {
		return mcfs.ErrDBRelatedUpdateFailed
	}

	return nil
}

// createDataFiles creates the datafiles entries for the datadirs_denorm table from
// the ids contained in a DataDir
func (d rDirs) createDataFiles(dataFileIDs []string) (dataFileEntries []schema.FileEntry, err error) {
	var errorReturn error
	for _, dataFileID := range dataFileIDs {
		var dataFile schema.File
		if err := model.Files.Q().ByID(dataFileID, &dataFile); err != nil {
			errorReturn = mcfs.ErrDBLookupFailed
			continue
		}

		dataFileEntry := schema.FileEntry{
			ID:        dataFile.ID,
			Name:      dataFile.Name,
			Owner:     dataFile.Owner,
			Birthtime: dataFile.Birthtime,
			Checksum:  dataFile.Checksum,
			Size:      dataFile.Size,
		}
		dataFileEntries = append(dataFileEntries, dataFileEntry)
	}

	return dataFileEntries, errorReturn
}

// RemoveFiles removes matching file ids from the directory and the dependent denorm
// table entries.
func (d rDirs) RemoveFiles(dir *schema.Directory, fileIDs ...string) error {
	dir.DataFiles = arrays.Strings.Remove(dir.DataFiles, fileIDs...)
	if err := d.Update(dir); err != nil {
		return err
	}
	var dirDenorm schema.DataDirDenorm
	if err := model.DirsDenorm.Q().ByID(dir.ID, &dirDenorm); err != nil {
		return mcfs.ErrDBRelatedUpdateFailed
	}
	dirDenorm.DataFiles = removeMatchingFileIDs(dirDenorm, fileIDs...)
	if err := model.DirsDenorm.Q().Update(dirDenorm.ID, dirDenorm); err != nil {
		return mcfs.ErrDBRelatedUpdateFailed
	}
	return nil
}

// removeMatchingFileIDs removes FileEntries from the list of entries that match that id.
func removeMatchingFileIDs(denorm schema.DataDirDenorm, fileIDs ...string) []schema.FileEntry {
	return denorm.Filter(func(f schema.FileEntry) bool {
		for _, fileID := range fileIDs {
			if fileID == f.ID {
				return false
			}
		}
		// Didn't find a match so keep entry
		return true
	})
}
