package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/base/model"
	"github.com/materials-commons/mcfs/mcd"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

func TestCreateDir(t *testing.T) {
	h := NewReqHandler(nil, "")
	h.user = "test@mc.org"

	// Test valid path

	createDirRequest := protocol.CreateDirReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		Path:      "Test/tdir1",
	}

	resp, err := h.createDir(&createDirRequest)

	if err != nil {
		t.Fatalf("Directory create failed with %s", err)
	}

	createdID := resp.ID
	var _ = createdID

	// Test existing directory

	resp, err = h.createDir(&createDirRequest)
	if err != nil {
		t.Fatalf("Create existing directory failed with %#v, err: %s", resp, err)
	}

	// Cleanup the created directory
	fmt.Println("Deleting datadir id:", createdID)
	model.Delete("datadirs", createdID, session)
	// Now cleanup the join table
	rv, _ := r.Table("project2datadir").GetAllByIndex("datadir_id", createdID).Delete().RunWrite(session)
	if rv.Deleted != 1 {
		t.Fatalf("Multiple entries in project2datadir matched. There should only have been one: %#v\n", rv)
	}

	// Test path outside of project
	createDirRequest.Path = "DIFFERENTPROJECT/tdir1"
	resp, err = h.createDir(&createDirRequest)
	if err == nil {
		t.Fatalf("Create dir outside of project succeeded %#v", resp)
	}

	// Test invalid project id
	createDirRequest.ProjectID = "abc123"
	createDirRequest.Path = "Test/tdir2"
	resp, err = h.createDir(&createDirRequest)
	if err == nil {
		t.Fatalf("Create dir with bad project succeeded %#v", resp)
	}

	// Test that fails if subdirs don't exist

	createDirRequest.ProjectID = "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3"
	createDirRequest.Path = "Test/tdir1/tdir2"

	resp, err = h.createDir(&createDirRequest)
	if err == nil {
		t.Fatalf("Create dir with missing subdirs succeeded %#v", resp)
	}
}
