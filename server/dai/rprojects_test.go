package dai

import (
	"fmt"
	"github.com/materials-commons/mcfs/base/db"
	"testing"
)

var _ = fmt.Println

func TestRProjectsByID(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")

	rprojs := newRProjects(session)

	// Test existing
	_, err := rprojs.ByID("9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3")
	if err != nil {
		t.Fatalf("Unable to retrieve existing project: %s", err)
	}
}

func TestRProjectsByName(t *testing.T) {
	rprojs := newRProjects(session)
	proj, err := rprojs.ByName("Test", "test@mc.org")
	if err != nil {
		t.Fatalf("Unable to find existing project 'Test', owner 'test@mc.org': %s", err)
	}

	var _ = proj
}

func TestRProjectsFiles(t *testing.T) {
	rprojs := newRProjects(session)
	files, err := rprojs.Files("9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", "")
	if err != nil {
		t.Fatalf("Unable to build list of files for existing project: %s", err)
	}
	if len(files) < 11 {
		t.Fatalf("Expected 11 entries, and got %d", len(files))
	}
}
