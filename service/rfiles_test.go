package service

import (
	"fmt"
	"github.com/materials-commons/base/db"
	"testing"
)

var _ = fmt.Println

func TestRFilesByID(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")
	rfiles := newRFiles()

	// Test existing
	f, err := rfiles.ByID("650ccb87-f423-499e-b644-2bb093eca86a")
	if err != nil {
		t.Fatalf("Unable retrieve existing file: %s", err)
	}

	var _ = f

	// Test non-existant
	_, err = rfiles.ByID("does-not-exist")
	if err == nil {
		t.Fatalf("Retrieved non-existant file does-not-exist")
	}
}
