package request

import (
	"fmt"
	//"github.com/materials-commons/mcfs/base/protocol"
	"testing"
)

var _ = fmt.Println

func TestProjectEntries(t *testing.T) {
	/*
	h := NewReqHandler(nil, "")
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
*/
}
