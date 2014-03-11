package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println
var _ = r.Table

func TestCreateDir(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "test@mc.org"

	// Test valid path

	createDirRequest := protocol.CreateDirReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		Path:      "Test/tdir1",
	}

	resp, status := h.createDir(&createDirRequest)

	if status != nil {
		t.Fatalf("Directory create failed with %s", status.err)
	}

	createdID := resp.ID
	var _ = createdID

	// Test existing directory

	resp, status = h.createDir(&createDirRequest)
	if status != nil {
		t.Fatalf("Create existing directory failed with %#v, err: %s", resp, status.err)
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
	resp, status = h.createDir(&createDirRequest)
	if status == nil {
		t.Fatalf("Create dir outside of project succeeded %#v", resp)
	}

	// Test invalid project id
	createDirRequest.ProjectID = "abc123"
	createDirRequest.Path = "Test/tdir2"
	resp, status = h.createDir(&createDirRequest)
	if status == nil {
		t.Fatalf("Create dir with bad project succeeded %#v", resp)
	}

	// Test that fails if subdirs don't exist

	createDirRequest.ProjectID = "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3"
	createDirRequest.Path = "Test/tdir1/tdir2"

	resp, status = h.createDir(&createDirRequest)
	if status == nil {
		t.Fatalf("Create dir with missing subdirs succeeded %#v", resp)
	}
}

func TestCreateProject(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "test@mc.org"

	createProjectRequest := protocol.CreateProjectReq{
		Name: "TestProject1__",
	}

	// Test create new project
	resp, status := h.createProject(&createProjectRequest)

	projectID := resp.ProjectID
	datadirID := resp.DataDirID

	if status != nil {
		t.Fatalf("Unable to create project")
	}

	// Make sure the created project is properly setup
	proj, err := model.GetProject(projectID, session)
	if err != nil {
		t.Errorf("Unable to retrieve project %s", projectID)
	}

	if proj.Name != "TestProject1__" {
		t.Errorf("Project Name not set")
	}

	if proj.DataDir == "" {
		t.Errorf("Project doesn't have a datadir associated with it")
	}

	// Make sure the join table is updated
	var p2d schema.Project2DataDir
	rql := r.Table("project2datadir").GetAllByIndex("project_id", projectID)
	err = model.GetRow(rql, session, &p2d)
	if err != nil {
		t.Errorf("Unable to find project in join table")
	}

	if p2d.DataDirID != datadirID {
		t.Errorf("Wrong datadir for project %#v expected %s", p2d, datadirID)
	}

	// Test create existing project
	resp, status = h.createProject(&createProjectRequest)
	if status.status != mc.ErrorCodeExists {
		t.Errorf("Creating an existing project should have returned err mc.ErrorCodeExists, returned %d instead", status.status)
	}

	// Delete before test so we can cleanup if there is a failure
	model.Delete("datadirs", datadirID, session)
	model.Delete("projects", projectID, session)
	model.Delete("project2datadir", p2d.ID, session)

	if status == nil {
		t.Fatalf("Created an existing project - shouldn't be able to")
	}

	if resp == nil {
		t.Fatalf("Creating an existing project should have returned its project id and datadir id")
	}

	if resp.ProjectID != projectID {
		t.Errorf("Creating an existing project returned the wrong project id")
	}

	if resp.DataDirID != datadirID {
		t.Errorf("Creating an existing project returned the wrong datadir id")
	}
	// Test create project with invalid name
	createProjectRequest.Name = "/InvalidName"
	resp, status = h.createProject(&createProjectRequest)
	if status == nil {
		t.Fatalf("Created project with Invalid name")
	}
}

func TestCreateFile(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "test@mc.org"

	// Test create with no size
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "testfile1.txt",
		Checksum:  "abc123",
	}

	resp, status := h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Created file with no size")
	}

	createFileRequest.Size = 1
	createFileRequest.Checksum = ""
	resp, status = h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Created file with no checksum")
	}

	// Test create a valid file
	createFileRequest.Size = 1
	createFileRequest.Checksum = "abc123"
	resp, status = h.createFile(&createFileRequest)
	if status != nil {
		t.Fatalf("Create file failed %s", status.err)
	}
	createdID := resp.ID

	// Validate the newly created datafile
	df, err := model.GetDataFile(createdID, session)
	if err != nil {
		t.Fatalf("Unable to retrieve a newly created datafile %s", err)
	}

	if df.Size != 1 {
		t.Fatalf("Wrong size %#v", df)
	}

	if df.Checksum != "abc123" {
		t.Fatalf("Bad checksum %#v", df)
	}

	if len(df.DataDirs) != 1 {
		t.Fatalf("Wrong number of datadirs %#v", df)
	}

	if df.DataDirs[0] != "f0ebb733-c75d-4983-8d68-242d688fcf73" {
		t.Fatalf("Wrong datadir inserted %#v", df)
	}

	if df.Access != "private" {
		t.Fatalf("Wrong access set %#v", df)
	}

	if df.Owner != "test@mc.org" {
		t.Fatalf("Wrong owner %#v", df)
	}

	if df.Name != "testfile1.txt" {
		t.Fatalf("Wrong name %#v", df)
	}

	// Test create new file that matches existing file size and checksum
	createFileRequest.DataDirID = "d0b001c6-fc0a-4e95-97c3-4427de68c0a5"
	resp, status = h.createFile(&createFileRequest)
	if status != nil {
		t.Fatalf("Unable to create file with matching size and checksum %s", status)
	}
	df, err = model.GetDataFile(resp.ID, session)
	if err != nil {
		t.Errorf("Unable to retrieve newly created datafile %s: %s", resp.ID, err)
	}
	if df.UsesID == "" {
		t.Errorf("UsesID is blank %#v", df)
	}

	if df.UsesID != createdID {
		t.Errorf("Wrong id for UsesID %#v", df)
	}

	createdID2 := resp.ID

	// Test creating an existing file
	resp, status = h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Allowed create of an existing file")
	}

	// Delete created files
	model.Delete("datafiles", createdID, session)
	model.Delete("datafiles", createdID2, session)

	// Test creating with an invalid project id
	validProjectID := createFileRequest.ProjectID
	createFileRequest.ProjectID = "abc123-doesnotexist"
	resp, status = h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Allowed creation of file with bad projectid")
	}

	// Test creating with an invalid datadir id
	createFileRequest.ProjectID = validProjectID
	createFileRequest.DataDirID = "abc123-doesnotexist"
	resp, status = h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Allowed creation of file with bad projectid")
	}

	// Test creating with a datadir not in project
	createFileRequest.DataDirID = "ae0cf23f-2588-4864-bf34-455b0aa23ed6"
	resp, status = h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Allowed creation of file in a datadir not in project")
	}

}
