package service

import (
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs"
)

// rDirs implements the Dirs interface for RethinkDB
type rDirs struct{}

// newRDirs create a new instance of rDirs
func newRDirs() rDirs {
	return rDirs{}
}

// ByID looks up a dir by its primary key. In RethinkDB this is the id field.
func (d rDirs) ByID(id string) (*schema.DataDir, error) {
	var dir schema.DataDir
	if err := model.Dirs.Q().ByID(id, &dir); err != nil {
		return nil, err
	}
	return &dir, nil
}

// Update updates an existing dir. If you are adding new files you should use the
// AddFiles method. This method will not update related items. AddFiles takes care
// of updating other related tables.
func (d rDirs) Update(dir *schema.DataDir) error {
	if err := model.Dirs.Q().Update(dir.ID, dir); err != nil {
		return err
	}
	return nil
}

// Insert creates a new dir. This method can return an error, with a valid
// DataDir. This happens when the dir is created, but one or more of the intermediate
// steps failed. The call needs to handle this case.
func (d rDirs) Insert(dir *schema.DataDir) (*schema.DataDir, error) {
	var newDir schema.DataDir
	if err := model.Dirs.Q().Insert(dir, &newDir); err != nil {
		return nil, mcfs.ErrDBInsertFailed
	}

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
func (d rDirs) AddFiles(dir *schema.DataDir, fileIDs ...string) error {
	// Add fileIds to DataDir
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
func (d rDirs) createDataFiles(dataFileIDs []string) (dataFileEntries []schema.DataFileEntry, err error) {
	var errorReturn error
	for _, dataFileID := range dataFileIDs {
		var dataFile schema.DataFile
		if err := model.Files.Q().ByID(dataFileID, &dataFile); err != nil {
			errorReturn = mcfs.ErrDBLookupFailed
			continue
		}

		dataFileEntry := schema.DataFileEntry{
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
