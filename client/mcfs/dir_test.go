package mcfs

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

var createDirTests = []struct {
	projectID   string
	projectName string
	path        string
	errorNil    bool
	description string
}{
	{"9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", "Test", "/tmp/abc", false, "Valid project bad path"},
	{"9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", "Test", "Test/abc", true, "Valid project path starts with project"},
	{"does not exist", "Test", "Test/abc.txt", false, "Valid project path bad project id"},
	{"9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", "Test", "/tmp/Test/abc", true, "Valid project full path containing project name"},
}

func TestCreateDir(t *testing.T) {

	for _, test := range createDirTests {
		_, err := c.CreateDir(test.projectID, test.projectName, test.path)
		switch {
		case err != nil && test.errorNil:
			t.Errorf("Expected error to be nil for test %s, err %s", test.description, err)
		case err == nil && !test.errorNil:
			t.Errorf("Expected err != nil for test %s", test.description)
		}
	}

	// Test creating an existing directory
	projID := "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3"
	projName := "Test"
	dirPath := "/tmp/Test/abc"
	dataDirID, err := c.CreateDir(projID, projName, dirPath)
	if err != nil {
		t.Errorf("Creating a directory that already exists returned wrong error code: %s", err)
	}
	if dataDirID == "" {
		t.Errorf("Creating an existing directory should have returned the id of the already created directory.")
	}
}
