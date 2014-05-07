package service

import (
	"fmt"
	"github.com/materials-commons/base/db"
	"testing"
)

var _ = fmt.Println

var rprojs = newRProjects()

func TestRProjectsByID(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")

	// Test existing
	_, err := rprojs.ByID("9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3")
	if err != nil {
		t.Fatalf("Unable to retrieve existing project: %s", err)
	}
}

func TestRProjectsFiles(t *testing.T) {
	files, err := rprojs.Files("9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", "")
	if err != nil {
		t.Fatalf("Unable to build list of files for existing project: %s", err)
	}
	if len(files) != 11 {
		t.Fatalf("Expected 11 entries, and got %d", len(files))
	}
}
