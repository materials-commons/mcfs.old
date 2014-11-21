package app

import (
	"testing"

	"github.com/materials-commons/gohandy/collections"
	"github.com/materials-commons/mcfs/testutils"
)

var tfs = NewFilesService(testutils.NewRFiles(),
	testutils.NewRDirs(),
	testutils.NewRProjects(),
	testutils.NewRGroups())

func TestFilesServiceCreateFile(t *testing.T) {
}

func TestFilesServiceValidate(t *testing.T) {

}

func TestFilesServiceCreateNewFile(t *testing.T) {
	file := File{
		Owner:       "test@mc.org",
		ProjectID:   "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DirectoryID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:        "newFile.txt",
		Checksum:    "abc123new",
		Size:        2,
	}

	// Create a new file and make sure it is setup correctly
	fn, _ := tfs.createNewFile(file)
	f, _ := tfs.files.ByID(fn.ID)

	f.Current = true
	tfs.files.Update(f)

	// The only thing to test at this point is Parent
	if f.Parent != "" {
		t.Errorf("Expected no parent and got one: %#v\n", f)
	}

	// Now create another version of the file. It should have f.Parent == f.ID
	// set to f.ID
	file.Name = "newFile.txt"
	file.Checksum = "abc123v2new"
	fn, _ = tfs.createNewFile(file)
	f2, _ := tfs.files.ByID(fn.ID)

	if f2.Parent != f.ID {
		t.Errorf("Expected new version of file to have its parent set to previous version.")
	}

	tfs.files.Delete(f.ID)
	tfs.files.Delete(f2.ID)
}

func TestFilesServiceNewFile(t *testing.T) {
	file := File{
		Owner:       "test@mc.org",
		ProjectID:   "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DirectoryID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:        "newFile.txt",
		Checksum:    "abc123",
		Size:        2,
	}

	// Test that parameters are setup correctly
	f := tfs.newFile(file)

	if f.Name != file.Name {
		t.Errorf("Wrong name %s/%s", f.Name, file.Name)
	}

	if f.Checksum != file.Checksum {
		t.Errorf("Wrong checksum %s/%s", f.Checksum, file.Checksum)
	}

	if f.Size != file.Size {
		t.Errorf("Wrong size %d/%d", f.Size, file.Size)
	}

	if f.Current != false {
		t.Errorf("Expected current to be false")
	}

	index := collections.Strings.Find(f.DataDirs, "f0ebb733-c75d-4983-8d68-242d688fcf73")
	if index == -1 {
		t.Errorf("Expected to find directory %s in list of directories %#v", "f0ebb733-c75d-4983-8d68-242d688fcf73", f.DataDirs)
	}

	// Insert file and then test to make sure we can find a duplicate when creating a new file.
	newf, _ := tfs.files.InsertEntry(f)

	file.Name = "newFilev2.txt"
	f = tfs.newFile(file)
	if f.UsesID != newf.ID {
		t.Errorf("Should have found a dup and set UsesID to it.")
	}

	tfs.files.Delete(newf.ID)
}
