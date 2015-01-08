package request

import (
	"fmt"
	"github.com/materials-commons/gohandy/collections"
	"github.com/materials-commons/mcfs/base/model"
	"github.com/materials-commons/mcfs/mcd/dai"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println

func TestCreateFile(t *testing.T) {
	h := NewReqHandler(nil, "")
	h.user = "test@mc.org"

	// Test create with no size
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "testfile1.txt",
		Checksum:  "abc123",
	}

	resp, err := h.createFile(&createFileRequest)
	if err == nil {
		t.Fatalf("Created file with no size")
	}

	createFileRequest.Size = 1
	createFileRequest.Checksum = ""
	resp, err = h.createFile(&createFileRequest)
	if err == nil {
		t.Fatalf("Created file with no checksum")
	}

	// Test create a valid file
	createFileRequest.Size = 1
	createFileRequest.Checksum = "abc123"
	resp, err = h.createFile(&createFileRequest)
	if err != nil {
		t.Fatalf("Create file failed %s", err)
	}
	createdID := resp.ID

	// Validate the newly created datafile
	df, err := model.GetFile(createdID, session)
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

	if df.Owner != "test@mc.org" {
		t.Fatalf("Wrong owner %#v", df)
	}

	if df.Name != "testfile1.txt" {
		t.Fatalf("Wrong name %#v", df)
	}

	// Test create new file that matches existing file size and checksum
	createFileRequest.DataDirID = "d0b001c6-fc0a-4e95-97c3-4427de68c0a5"
	resp, err = h.createFile(&createFileRequest)
	if err != nil {
		t.Fatalf("Unable to create file with matching size and checksum %s", err)
	}
	df, err = model.GetFile(resp.ID, session)
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
	resp, err = h.createFile(&createFileRequest)
	if err != nil {
		t.Fatalf("Failed creating an existing file")
	}

	// Delete created files
	model.Delete("datafiles", createdID, session)
	model.Delete("datafiles", createdID2, session)

	// Test creating with an invalid project id
	validProjectID := createFileRequest.ProjectID
	createFileRequest.ProjectID = "abc123-doesnotexist"
	resp, err = h.createFile(&createFileRequest)
	if err == nil {
		t.Fatalf("Allowed creation of file with bad projectid")
	}

	// Test creating with an invalid datadir id
	createFileRequest.ProjectID = validProjectID
	createFileRequest.DataDirID = "abc123-doesnotexist"
	resp, err = h.createFile(&createFileRequest)
	if err == nil {
		t.Fatalf("Allowed creation of file with bad projectid")
	}

	// Test creating with a datadir not in project
	createFileRequest.DataDirID = "ae0cf23f-2588-4864-bf34-455b0aa23ed6"
	resp, err = h.createFile(&createFileRequest)
	if err == nil {
		t.Fatalf("Allowed creation of file in a datadir not in project")
	}
}

func TestNewFile(t *testing.T) {
	cfh := newCreateFileHandler("test@mc.org", dai.New(dai.RethinkDB))

	req := &protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "newFile.txt",
		Checksum:  "abc123",
		Size:      2,
	}

	// Test that parameters are setup correctly
	f := cfh.newFile(req)

	if f.Name != req.Name {
		t.Errorf("Wrong name %s/%s", f.Name, req.Name)
	}

	if f.Checksum != req.Checksum {
		t.Errorf("Wrong checksum %s/%s", f.Checksum, req.Checksum)
	}

	if f.Size != req.Size {
		t.Errorf("Wrong size %d/%d", f.Size, req.Size)
	}

	if f.Current != false {
		t.Errorf("Expected current to be false")
	}

	index := collections.Strings.Find(f.DataDirs, "f0ebb733-c75d-4983-8d68-242d688fcf73")
	if index == -1 {
		t.Errorf("Expected to find directory %s in list of directories %#v", "f0ebb733-c75d-4983-8d68-242d688fcf73", f.DataDirs)
	}

	// Insert file and then test to make sure we can find a duplicate when creating a new file.
	newf, _ := cfh.dai.File.InsertEntry(f)

	req.Name = "newFilev2.txt"
	f = cfh.newFile(req)
	if f.UsesID != newf.ID {
		t.Errorf("Should have found a dup and set UsesID to it.")
	}

	cfh.dai.File.Delete(newf.ID)
}

func TestCreateNewFile(t *testing.T) {
	cfh := newCreateFileHandler("test@mc.org", dai.New(dai.RethinkDB))

	req := &protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "newFile.txt",
		Checksum:  "abc123new",
		Size:      2,
	}

	// Create a new file and make sure it is setup correctly
	resp, _ := cfh.createNewFile(req)
	f, _ := cfh.dai.File.ByID(resp.ID)

	f.Current = true
	cfh.dai.File.Update(f)

	// The only thing to test at this point is Parent
	if f.Parent != "" {
		t.Errorf("Expected no parent and got one: %#v\n", f)
	}

	// Now create another version of the file. It should have f.Parent == f.ID
	// set to f.ID
	req.Name = "newFile.txt"
	req.Checksum = "abc123v2new"
	resp, _ = cfh.createNewFile(req)
	f2, _ := cfh.dai.File.ByID(resp.ID)

	if f2.Parent != f.ID {
		t.Errorf("Expected new version of file to have its parent set to previous version.")
	}

	cfh.dai.File.Delete(f.ID)
	cfh.dai.File.Delete(f2.ID)
}
