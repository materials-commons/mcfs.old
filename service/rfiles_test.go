package service

import (
	"fmt"
	"github.com/materials-commons/base/db"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/schema"
	"testing"
)

var _ = fmt.Println

func TestRFilesByID(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")
	rfiles := newRFiles()

	// Test existing
	_, err := rfiles.ByID("650ccb87-f423-499e-b644-2bb093eca86a")
	if err != nil {
		t.Fatalf("Unable retrieve existing file: %s", err)
	}

	// Test non-existant
	_, err = rfiles.ByID("does-not-exist")
	if err == nil {
		t.Fatalf("Retrieved non-existant file does-not-exist")
	}
}

func TestRFilesByPath(t *testing.T) {
	rfiles := newRFiles()

	// Lookup an existing file that exists in the given directory
	f, err := rfiles.ByPath("2H-10X2.JPG", "d0b001c6-fc0a-4e95-97c3-4427de68c0a5")
	if err != nil {
		t.Fatalf("Unable to lookup existing file in existing directory: %s", err)
	}

	if f == nil {
		t.Fatalf("Lookup succeeded, but returned no entry")
	}

	// Lookup an file in the wrong directory
	_, err = rfiles.ByPath("2H-10X2.JPG", "c3d72271-4a32-4080-a6a3-b4c6a5c4b986")
	if err == nil {
		t.Fatalf("Found file in a directory that it is not in")
	}

	// Lookup an file in a non-existent directory
	_, err = rfiles.ByPath("2H-10X2.JPG", "dir-does-not-exist")
	if err == nil {
		t.Fatalf("Found file in a non-existent directory")
	}

	// Lookup a non-existent file in a non-existent directory
	f, err = rfiles.ByPath("file-does-not-exist", "dir-does-not-exist")
	if err == nil {
		t.Fatalf("No error when looking up non-existent file in non-existent directory")
	}

	if f != nil {
		t.Fatalf("Found non-existent file in non-existent directory")
	}

	// Insert a file that doesn't have current set and then see if we can find it.
	file := schema.NewFile("testfile.test", "test@mc.org")
	file.Current = false
	var nf *schema.File
	file.DataDirs = append(file.DataDirs, "c3d72271-4a32-4080-a6a3-b4c6a5c4b986")
	nf, err = rfiles.Insert(&file)
	if err != nil {
		t.Fatalf("Unable to insert file testfile.test")
	}

	_, err = rfiles.ByPath("testfile.test", "c3d72271-4a32-4080-a6a3-b4c6a5c4b986")
	if err != mc.ErrNotFound {
		t.Fatalf("Lookup of testfile.test should have returned err 'mc.ErrNotFound', returned %s instead", err)
	}

	rfiles.Delete(nf.ID)
}

func TestRFilesByChecksum(t *testing.T) {
	rfiles := newRFiles()

	// Lookup an existing checksum
	f, err := rfiles.ByChecksum("72d47a675e81cf4a283aaf67587ddd28")
	if err != nil {
		t.Fatalf("Failed looking up an existing checksum: %s", err)
	}

	if f == nil {
		t.Fatalf("No error, but also didn't return a file entry")
	}

	// Lookup an non-existent checksum
	f, err = rfiles.ByChecksum("does-not-exist")
	if err == nil {
		t.Fatalf("No error returned when looking up a file by a bad checksum")
	}

	if f != nil {
		t.Fatalf("Found a file for a bad checksum")
	}
}

func TestRFilesInsert(t *testing.T) {
	// Insert a new item
	dataFile := schema.NewFile("testfile.txt", "test@mc.org")
	dataFile.DataDirs = append(dataFile.DataDirs, "d0b001c6-fc0a-4e95-97c3-4427de68c0a5")
	rfiles := newRFiles()
	newDF, err := rfiles.Insert(&dataFile)
	if err != nil {
		t.Fatalf("Unable to insert new datafile: %s", err)
	}

	// Now test that the denorm table was properly updated

	rfilesCleanup(newDF)

	// Insert with an existing id, should fail
}

func rfilesCleanup(f *schema.File) {
	fmt.Println("Deleting file: ", f.ID, f.Name)
	rf := newRFiles()
	rf.Delete(f.ID)
}
