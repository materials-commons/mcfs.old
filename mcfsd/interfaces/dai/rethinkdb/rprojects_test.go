package rethinkdb

import (
	"fmt"
	"testing"

	"github.com/materials-commons/mcfs/testutils/tdb"
)

var _ = fmt.Println

func TestRProjectsByID(t *testing.T) {
	rprojs := NewRProjects(tdb.NewSession())

	// Test existing
	_, err := rprojs.ByID("9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3")
	if err != nil {
		t.Fatalf("Unable to retrieve existing project: %s", err)
	}
}

func TestRProjectsByName(t *testing.T) {
	rprojs := NewRProjects(tdb.NewSession())
	proj, err := rprojs.ByName("Test", "test@mc.org")
	if err != nil {
		t.Fatalf("Unable to find existing project 'Test', owner 'test@mc.org': %s", err)
	}

	var _ = proj
}

func TestRProjectsFiles(t *testing.T) {
	rprojs := NewRProjects(tdb.NewSession())
	files, err := rprojs.Files("9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", "")
	if err != nil {
		t.Fatalf("Unable to build list of files for existing project: %s", err)
	}
	if len(files) < 11 {
		t.Fatalf("Expected 11 entries, and got %d", len(files))
	}
}
