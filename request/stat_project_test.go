package request

import (
	"fmt"
	"github.com/materials-commons/base/db"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println

var (
	projectNameTests = []struct {
		name        string
		errorNil    bool
		description string
	}{
		{"Test", true, "Test existing project with access"},
		{"Test2", false, "Testing existing project without access"},
		{"Does not exist", false, "Test non existant project"},
	}
)

func TestGetProjectByName(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")
	p := &statProjectHandler{
		session: session,
		user:    "test@mc.org",
	}

	for _, test := range projectNameTests {
		_, err := p.getProjectByName(test.name)
		switch {
		case err != nil && test.errorNil:
			t.Errorf("Expected error to be nil for test %s, err %s", test.description, err)
		case err == nil && !test.errorNil:
			t.Errorf("Expected err != nil for test %s", test.description)
		}
	}
}

func TestGetProjectEntries(t *testing.T) {
	p := &statProjectHandler{
		session: session,
		user:    "test@mc.org",
	}

	// Test existing project
	projectID := "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3"
	results, err := p.projectDirList(projectID, "")
	if err != nil {
		t.Errorf("Query on known project id failed %s", err)
	}

	if len(results) == 0 {
		t.Errorf("Expected # of results to be greater than 0")
	}

	// Test bad id
	results, err = p.projectDirList("bad-id", "")
	if err != mc.ErrNotFound {
		t.Errorf("Error not equal to ErrNotFound for bad project id: %s", err)
	}

	if results != nil {
		t.Errorf("Expected results to be nil")
	}
}

func TestProjectEntries(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "test@mc.org"

	req := protocol.StatProjectReq{
		Name: "Test",
	}

	// Test project we have access to
	resp, err := h.statProject(&req)
	if err != nil {
		t.Errorf("Unable to access project I own: %s", err)
	}

	if resp.ProjectID != "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3" {
		t.Errorf("Bad project id returned %#v\n", resp)
	}

	if len(resp.Entries) == 0 {
		t.Errorf("No entries for project %s %#v\n", req.Name, resp)
	}

	// Test bad project name that doesn't exist
	req.Name = "Does-Not-Exist"
	resp, err = h.statProject(&req)
	if err == nil {
		t.Errorf("No error for bad project")
	}

	if resp != nil {
		t.Errorf("Bad project name should have nil resp %#v", resp)
	}

	// Test project name that we don't have access to
	req.Name = "Test2"
	if err == nil {
		t.Errorf("Got access to project I don't have permissions on")
	}

	if resp != nil {
		t.Errorf("Project without access should have nil resp %#v", resp)
	}
}
