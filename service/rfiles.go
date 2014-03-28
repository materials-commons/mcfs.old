package service

import (
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs"
)

// rFiles implements the Files interface for RethinkDB
type rFiles struct{}

// newRFiles creates a new instance of rFiles
func newRFiles() rFiles {
	return rFiles{}
}

// ByID looks up a file by its primary key. In RethinkDB this is the id field.
func (f rFiles) ByID(id string) (*schema.File, error) {
	var file schema.File
	if err := model.Files.Q().ByID(id, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// Update updates an existing datafile. If you are adding the datafile to a directory
// you should use the AddDirectories method. This method will not update related items.
func (f rFiles) Update(file *schema.File) error {
	if err := model.Files.Q().Update(file.ID, file); err != nil {
		return err
	}
	return nil
}

// Insert creates a new file entry.
func (f rFiles) Insert(file *schema.File) (*schema.File, error) {
	var newFile schema.File
	if err := model.Files.Q().Insert(file, &newFile); err != nil {
		return nil, err
	}
	if err := f.insertIntoDenorm(&newFile); err != nil {
		return &newFile, err
	}
	return &newFile, nil
}

// Delete deletes a file. It updates dependent objects.
func (f rFiles) Delete(id string) error {
	file, err := f.ByID(id)
	if err != nil {
		return err
	}

	if err := model.Files.Q().Delete(id); err != nil {
		return err
	}

	rDirs := newRDirs()
	for _, dirID := range file.DataDirs {
		d, _ := rDirs.ByID(dirID)
		rDirs.RemoveFiles(d, file.ID)
	}
	return nil
}

// AddDirectories adds new directories to a file. It updates all related items
// and join tables.
func (f rFiles) AddDirectories(file *schema.File, dirIDs ...string) error {
	// Add directories to to datafile
	for _, id := range dirIDs {
		file.DataDirs = append(file.DataDirs, id)
	}

	if err := model.Files.Q().Update(file.ID, file); err != nil {
		return mcfs.ErrDBUpdateFailed
	}

	// Add entries to the denorm table for this file
	return f.insertIntoDenorm(file)
}

// insertIntoDenrom updates the denorm table with the new file entries.
func (f rFiles) insertIntoDenorm(file *schema.File) error {
	fileEntry := schema.FileEntry{
		ID:        file.ID,
		Name:      file.Name,
		Owner:     file.Owner,
		Birthtime: file.Birthtime,
		Checksum:  file.Checksum,
		Size:      file.Size,
	}

	for _, ddirID := range file.DataDirs {
		var ddirDenorm schema.DataDirDenorm
		if err := model.DirsDenorm.Q().ByID(ddirID, &ddirDenorm); err != nil {
			return err
		}

		ddirDenorm.DataFiles = append(ddirDenorm.DataFiles, fileEntry)
		if err := model.DirsDenorm.Q().Update(ddirID, ddirDenorm); err != nil {
			return err
		}
	}
	return nil
}
