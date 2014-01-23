package request

import (
	"fmt"
	"github.com/materials-commons/contrib/mc"
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
		{"Test_Proj", true, "Test existing project with access"},
		{"Synthetic Tooth", false, "Testing existing project without access"},
		{"Does not exist", false, "Test non existant project"},
	}
)

func TestGetProjectByName(t *testing.T) {
	p := &projectEntryHandler{
		session: session,
		user:    "gtarcea@umich.edu",
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
	p := &projectEntryHandler{
		session: session,
		user:    "gtarcea@umich.edu",
	}

	// Test existing project
	projectID := "904886a7-ea57-4de7-8125-6e18c9736fd0"
	results, err := p.getProjectEntries(projectID)
	if err != nil {
		t.Errorf("Query on known project id failed %s", err)
	}

	if len(results) == 0 {
		t.Errorf("Expected # of results to be greater than 0")
	}

	// Test bad id
	results, err = p.getProjectEntries("bad-id")
	if err != mc.ErrNotFound {
		t.Errorf("Error not equal to ErrNotFound for bad project id")
	}

	if results != nil {
		t.Errorf("Expected results to be nil")
	}
}

func TestProjectEntries(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "gtarcea@umich.edu"

	req := protocol.ProjectEntriesReq{
		Name: "Test_Proj",
	}

	// Test project we have access to
	resp, err := h.projectEntries(&req)
	if err != nil {
		t.Errorf("Unable to access project I own: %s", err)
	}

	if resp.ProjectID != "c33edab7-a65f-478e-9fa6-9013271c73ea" {
		t.Errorf("Bad project id returned %#v\n", resp)
	}

	if len(resp.Entries) == 0 {
		t.Errorf("No entries for project %s %#v\n", req.Name, resp)
	}

	// Test bad project name that doesn't exist
	req.Name = "Does-Not-Exist"
	resp, err = h.projectEntries(&req)
	if err == nil {
		t.Errorf("No error for bad project")
	}

	if resp != nil {
		t.Errorf("Bad project name should have nil resp %#v", resp)
	}

	// Test project name that we don't have access to
	req.Name = "Synthetic Tooth"
	if err == nil {
		t.Errorf("Got access to project I don't have permissions on")
	}

	if resp != nil {
		t.Errorf("Project without access should have nil resp %#v", resp)
	}
}
