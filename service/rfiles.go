package service

import (
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
)

// rFiles implements the Files interface for RethinkDB
type rFiles struct{}

// newRFiles creates a new instance of rFiles
func newRFiles() rFiles {
	return rFiles{}
}

// ByID looks up a file by its primary key. In RethinkDB this is the id field.
func (f rFiles) ByID(id string) (*schema.DataFile, error) {
	var file schema.DataFile
	if err := model.Files.Q().ByID(id, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// Update updates an existing datafile. If you are adding the datafile to a directory
// you should use the AddDirectories method. This method will not update related items.
func (f rFiles) Update(file *schema.DataFile) error {
	if err := model.Files.Q().Update(file.ID, file); err != nil {
		return err
	}
	return nil
}

// Insert creates a new file entry.
func (f rFiles) Insert(file *schema.DataFile) (*schema.DataFile, error) {
	// TODO: Update denorm table with new file for directory
	var newFile schema.DataFile
	if err := model.Files.Q().Insert(file, &newFile); err != nil {
		return nil, err
	}
	return &newFile, nil
}

func (f rFiles) AddDirectories(file *schema.DataFile, dirIDs ...string) error {
	return nil
}
