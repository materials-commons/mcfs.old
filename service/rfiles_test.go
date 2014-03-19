package service

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/db"
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

func TestRFilesInsert(t *testing.T) {
	// Insert a new item
	dataFile := schema.NewFile("testfile.txt", "private", "test@mc.org")
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
	session, _ := db.RSession()
	r.Table("datafiles").Get(f.ID).Delete().RunWrite(session)
}
