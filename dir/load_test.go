package dir

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestLoad(t *testing.T) {
	d, err := Load("testdir")
	if err != nil {
		t.Fatalf("Failed to create directory for testdir: %s", err)
	}

	if len(d.Files) != 4 {
		t.Fatalf("Wrong length of entries in testdir Files, expected 4, got %d", len(d.Files))
	}

	if len(d.SubDirectories) != 2 {
		t.Fatalf("Wrong length of entries in testdir subdirectories, expected 2, got %d", len(d.SubDirectories))
	}
}
