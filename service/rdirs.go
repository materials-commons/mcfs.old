package service

import (
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
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

// Insert creates a new dir.
func (d rDirs) Insert(dir *schema.DataDir) (*schema.DataDir, error) {
	var newDir schema.DataDir
	if err := model.Dirs.Q().Insert(dir, &newDir); err != nil {
		return nil, err
	}

	var ddirDenorm = schema.DataDirDenorm{
		ID:        newDir.ID,
		Name:      newDir.Name,
		Owner:     newDir.Owner,
		Birthtime: newDir.Birthtime,
	}

	if len(newDir.DataFiles) > 0 {
		ddirDenorm.DataFiles = d.createDataFiles(newDir.DataFiles)
	}

	if err := model.DirsDenorm.Q().Insert(ddirDenorm, nil); err != nil {
		// Ack, database out of sync!
	}

	return &newDir, nil
}

// AddFiles adds new file ids to a dir. It updates all related items and join tables.
func (d rDirs) AddFiles(dir *schema.DataDir, fileIDs ...string) error {
	// Add fileIds to DataDir
	for _, id := range fileIDs {
		dir.DataFiles = append(dir.DataFiles, id)
	}

	if err := model.Dirs.Q().Update(dir.ID, dir); err != nil {
		return nil
	}

	// Add entries to the denorm table for this dir.
	var dirDenorm schema.DataDirDenorm
	newEntries := d.createDataFiles(fileIDs)
	if err := model.DirsDenorm.Q().ByID(dir.ID, &dirDenorm); err != nil {
		// TODO: What to do?
	}

	dirDenorm.DataFiles = append(dirDenorm.DataFiles, newEntries...)
	if err := model.DirsDenorm.Q().Update(dirDenorm.ID, dirDenorm); err != nil {
		// TODO: What to do?
	}

	return nil
}

// createDataFiles creates the datafiles entries for the datadirs_denorm table from
// the ids contained in a DataDir
func (d rDirs) createDataFiles(dataFileIDs []string) (dataFileEntries []schema.DataFileEntry) {
	for _, dataFileID := range dataFileIDs {
		var dataFile schema.DataFile
		if err := model.Files.Q().ByID(dataFileID, &dataFile); err != nil {
			// TODO: How to handle this type of error?
			// The database will be out of sync, but it isn't clear what
			// we should do. For now log error and go on.
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
	return dataFileEntries
}
