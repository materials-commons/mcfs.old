package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/contrib/mc"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println
var _ = r.Table

func TestCreateDir(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "gtarcea@umich.edu"

	// Test valid path

	createDirRequest := protocol.CreateDirReq{
		ProjectID: "904886a7-ea57-4de7-8125-6e18c9736fd0",
		Path:      "WE43 Heat Treatments/tdir1",
	}

	resp, status := h.createDir(&createDirRequest)

	if status != nil {
		t.Fatalf("Directory create failed with %s", status.err)
	}

	createdId := resp.ID
	var _ = createdId

	// Test existing directory

	resp, status = h.createDir(&createDirRequest)
	if status != nil {
		t.Fatalf("Create existing directory failed with %#v, err: %s", resp, status.err)
	}

	// Cleanup the created directory
	fmt.Println("Deleting datadir id:", createdId)
	model.Delete("datadirs", createdId, session)
	// Now cleanup the join table
	rv, _ := r.Table("project2datadir").GetAllByIndex("datadir_id", createdId).Delete().RunWrite(session)
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
	createDirRequest.Path = "WE43 Heat Treatments/tdir2"
	resp, status = h.createDir(&createDirRequest)
	if status == nil {
		t.Fatalf("Create dir with bad project succeeded %#v", resp)
	}

	// Test that fails if subdirs don't exist

	createDirRequest.ProjectID = "904886a7-ea57-4de7-8125-6e18c9736fd0"
	createDirRequest.Path = "WE43 Heat Treatments/tdir1/tdir2"

	resp, status = h.createDir(&createDirRequest)
	if status == nil {
		t.Fatalf("Create dir with missing subdirs succeeded %#v", resp)
	}
}

func TestCreateProject(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "gtarcea@umich.edu"

	createProjectRequest := protocol.CreateProjectReq{
		Name: "TestProject1__",
	}

	// Test create new project
	resp, status := h.createProject(&createProjectRequest)

	projectId := resp.ProjectID
	datadirId := resp.DataDirID

	if status != nil {
		t.Fatalf("Unable to create project")
	}

	// Make sure the created project is properly setup
	proj, err := model.GetProject(projectId, session)
	if err != nil {
		t.Errorf("Unable to retrieve project %s", projectId)
	}

	if proj.Name != "TestProject1__" {
		t.Errorf("Project Name not set")
	}

	if proj.DataDir == "" {
		t.Errorf("Project doesn't have a datadir associated with it")
	}

	// Make sure the join table is updated
	var p2d Project2Datadir
	rql := r.Table("project2datadir").GetAllByIndex("project_id", projectId)
	err = model.GetRow(rql, session, &p2d)
	if err != nil {
		t.Errorf("Unable to find project in join table")
	}

	if p2d.DataDirID != datadirId {
		t.Errorf("Wrong datadir for project %#v expected %s", p2d, datadirId)
	}

	// Test create existing project
	resp, status = h.createProject(&createProjectRequest)
	if status.status != mc.ErrorCodeExists {
		t.Errorf("Creating an existing project should have returned err mc.ErrorCodeExists, returned %d instead", status.status)
	}

	// Delete before test so we can cleanup if there is a failure
	model.Delete("datadirs", datadirId, session)
	model.Delete("projects", projectId, session)
	model.Delete("project2datadir", p2d.Id, session)

	if status == nil {
		t.Fatalf("Created an existing project - shouldn't be able to")
	}

	if resp == nil {
		t.Fatalf("Creating an existing project should have returned its project id and datadir id")
	}

	if resp.ProjectID != projectId {
		t.Errorf("Creating an existing project returned the wrong project id")
	}

	if resp.DataDirID != datadirId {
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
	h.user = "gtarcea@umich.edu"

	// Test create with no size
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "c33edab7-a65f-478e-9fa6-9013271c73ea",
		DataDirID: "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys_Data",
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
	createdId := resp.ID

	// Validate the newly created datafile
	df, err := model.GetDataFile(createdId, session)
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

	if df.DataDirs[0] != "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys_Data" {
		t.Fatalf("Wrong datadir inserted %#v", df)
	}

	if df.Access != "private" {
		t.Fatalf("Wrong access set %#v", df)
	}

	if df.Owner != "gtarcea@umich.edu" {
		t.Fatalf("Wrong owner %#v", df)
	}

	if df.Name != "testfile1.txt" {
		t.Fatalf("Wrong name %#v", df)
	}

	// Test create new file that matches existing file size and checksum
	createFileRequest.DataDirID = "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys"
	resp, status = h.createFile(&createFileRequest)
	if status != nil {
		t.Errorf("Unable to create file with matching size and checksum")
	}
	df, err = model.GetDataFile(resp.ID, session)
	if err != nil {
		t.Errorf("Unable to retrieve newly created datafile %s: %s", resp.ID, err)
	}
	if df.UsesID == "" {
		t.Errorf("UsesID is blank %#v", df)
	}

	if df.UsesID != createdId {
		t.Errorf("Wrong id for UsesID %#v", df)
	}

	createdId2 := resp.ID

	// Test creating an existing file
	resp, status = h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Allowed create of an existing file")
	}

	// Delete created files
	model.Delete("datafiles", createdId, session)
	model.Delete("datafiles", createdId2, session)

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
	createFileRequest.DataDirID = "mcfada@umich.edu$Synthetic Tooth_Presentation_MCubed"
	resp, status = h.createFile(&createFileRequest)
	if status == nil {
		t.Fatalf("Allowed creation of file in a datadir not in project")
	}

}
