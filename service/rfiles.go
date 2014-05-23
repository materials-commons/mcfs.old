package service

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/gohandy/collections"
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

// ByPath looks up a file by its name in a specific directory. It only returns the
// current file, not hidden files.
func (f rFiles) ByPath(name, dirID string) (*schema.File, error) {
	rql := model.Files.T().GetAllByIndex("name", name).
		Filter(r.Row.Field("datadirs").Contains(dirID).And(r.Row.Field("current").Eq(true)))
	var file schema.File
	if err := model.Files.Q().Row(rql, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// ByPathPartials returns all the partials matching name in the given directory. A
// partial is a file that has not completed uploading. A file is a partial when its
// size is not equal to uploaded.
func (f rFiles) ByPathPartials(name, dirID string) ([]schema.File, error) {
	rql := model.Files.T().GetAllByIndex("name", name).
		Filter(r.Row.Field("datadirs").Contains(dirID).
		And(r.Row.Field("uploaded").Ne(r.Row.Field("size"))))
	var files []schema.File
	if err := model.Files.Q().Rows(rql, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// ByPathChecksum looks up a file by its name, checksum and directory. This method
// can return files that are only partially uploaded. It can return
func (f rFiles) ByPathChecksum(name, dirID, checksum string) ([]schema.File, error) {
	var files []schema.File
	rql := model.Files.T().GetAllByIndex("name", name).
		Filter(r.Row.Field("datadirs").Contains(dirID).
		And(r.Row.Field("checksum").Eq(checksum)))
	if err := model.Files.Q().Rows(rql, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// ByChecksum looks up a file by its checksum. This routine only returns the original
// root entry, it will not return entries that are duplicates and point at the root.
func (f rFiles) ByChecksum(checksum string) (*schema.File, error) {
	rql := model.Files.T().GetAllByIndex("checksum", checksum).Filter(r.Row.Field("usesid").Eq(""))
	var file schema.File
	if err := model.Files.Q().Row(rql, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

// MatchOn looks up files by key.
func (f rFiles) MatchOn(key, value string) ([]schema.File, error) {
	var files []schema.File
	rql := model.Files.T().GetAllByIndex(key, value)
	if err := model.Files.Q().Rows(rql, &files); err != nil {
		return nil, err
	}
	return files, nil
}

// Hide keeps the file around, but removes it from all dependent objects. This allows
// multiple versions of a file to exist, but only the current version to be used.
func (f rFiles) Hide(file *schema.File) error {
	file.Current = false
	model.Files.Q().Update(file.ID, file)
	return f.removeFromDependents(file)
}

// Update updates an existing datafile. If you are adding the datafile to a directory
// you should use the AddDirectories method. This method will not update related items.
func (f rFiles) Update(file *schema.File) error {
	if err := model.Files.Q().Update(file.ID, file); err != nil {
		return err
	}
	return nil
}

// Insert creates a new file entry. Insert updates the directory and other
// dependent objects in the system.
func (f rFiles) Insert(file *schema.File) (*schema.File, error) {
	var newFile schema.File
	if err := model.Files.Q().Insert(file, &newFile); err != nil {
		return nil, err
	}
	if err := f.AddDirectories(&newFile, file.DataDirs...); err != nil {
		return &newFile, err
	}
	return &newFile, nil
}

// InsertEntry creates a new file entry. It does not update any dependent
// objects. You can use AddDirectory to add/update the directory this file
// belongs in. This method exists to allow for the creation of new file
// objects that will be linked into the rest of the system at a late date.
func (f rFiles) InsertEntry(file *schema.File) (*schema.File, error) {
	var newFile schema.File
	if err := model.Files.Q().Insert(file, &newFile); err != nil {
		return nil, err
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

	return f.removeFromDependents(file)
}

// removeFromDependents removes the file from all the other objects in
// the database that refer to it.
func (f rFiles) removeFromDependents(file *schema.File) error {
	// Need to delete file from dependent objects
	rdirs := newRDirs()
	var rv error
	for _, dirID := range file.DataDirs {
		ddir, err := rdirs.ByID(dirID)
		if err != nil {
			rv = mcfs.ErrDBRelatedUpdateFailed
		}

		err = rdirs.RemoveFiles(ddir, file.ID)
		if err != nil {
			rv = mcfs.ErrDBRelatedUpdateFailed
		}
	}

	return rv
}

// AddDirectories adds new directories to a file. It updates all related items
// and join tables.
func (f rFiles) AddDirectories(file *schema.File, dirIDs ...string) error {
	rdirs := newRDirs()
	var rv error
	for _, ddirID := range dirIDs {
		if index := collections.Strings.Find(file.DataDirs, ddirID); index == -1 {
			file.DataDirs = append(file.DataDirs, ddirID)
		}
		dir, err := rdirs.ByID(ddirID)
		rdirs.AddFiles(dir, file.ID)
		if err != nil {
			rv = mcfs.ErrDBRelatedUpdateFailed
		}
	}
	f.Update(file)
	return rv
}
