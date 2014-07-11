package materials

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

var (
	expectedTop = ProjectFileEntry{
		ID:          "0",
		ParentID:    "",
		Path:        "/tmp/tproj",
		HrefPath:    "",
		DisplayName: "/tmp/tproj",
		Type:        "datadir",
	}

	expectedChildDir = ProjectFileEntry{
		ID:          "1",
		ParentID:    "0",
		Path:        "/tmp/tproj/a",
		HrefPath:    "",
		DisplayName: "a",
		Type:        "datadir",
	}

	expectedDF = ProjectFileEntry{
		ID:          "2",
		ParentID:    "1",
		Path:        "/tmp/tproj/a/a.txt",
		HrefPath:    "tproj/a/a.txt",
		DisplayName: "a.txt",
		Type:        "datafile",
	}
)

func TestBuildTree(t *testing.T) {
	projects, _ := OpenProjectDB(testData)
	tproj, _ := projects.Find("tproj")
	tree, err := tproj.Tree()
	if err != nil {
		t.Fatalf("Creating tree got unexpected error %s\n", err.Error())
	}

	if len(tree) != 1 {
		t.Fatalf("Expected tree length 1, got %d\n", len(tree))
	}

	compare(*tree[0], expectedTop, 1, t)

	if len(tree[0].Children) != 1 {
		t.Fatalf("Expected tree[0].Children length 1, got %d\n", len(tree[0].Children))
	}

	child := tree[0].Children[0]
	compare(*child, expectedChildDir, 1, t)

	if len(tree[0].Children[0].Children) != 1 {
		t.Fatalf("Expected tree[0].Children[0].Children length 1, got %d\n", len(tree[0].Children[0].Children))
	}

	child = tree[0].Children[0].Children[0]
	compare(*child, expectedDF, 0, t)
}

func TestBuildTreeBadProj(t *testing.T) {
	projects, _ := OpenProjectDB(testData)
	badProj, _ := projects.Find("proj 2")
	tree, err := badProj.Tree()
	if err == nil {
		t.Fatalf("Expected an error, got nil instead\n")
	}

	if len(tree) != 0 {
		t.Fatalf("Expected a zero length tree, got %d instead", len(tree))
	}
}

func compare(child, expected ProjectFileEntry, clen int, t *testing.T) {
	if child.ID != expected.ID {
		t.Fatalf("Ids not equal, expected %s, got %s", expected.ID, child.ID)
	}

	if child.ParentID != expected.ParentID {
		t.Fatalf("ParentIds not equal, expected %s, got %s", expected.ParentID, child.ParentID)
	}

	if child.Path != expected.Path {
		t.Fatalf("Paths not equal, expected %s, got %s", expected.Path, child.Path)
	}

	if child.HrefPath != expected.HrefPath {
		t.Fatalf("HrefPaths not equal, expected %s, got %s", expected.HrefPath, child.HrefPath)
	}

	if child.DisplayName != expected.DisplayName {
		t.Fatalf("DisplayNames not equal, expected %s, got %s", expected.DisplayName, child.DisplayName)
	}

	if child.Type != expected.Type {
		t.Fatalf("Types not equal, expected %s, got %s", expected.Type, child.Type)
	}

	if len(child.Children) != clen {
		t.Fatalf("Children length not equal, expected %d, got %d", clen, len(child.Children))
	}
}
