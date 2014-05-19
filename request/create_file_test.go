package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println
var _ = r.Table

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

	// There should now be 2 entries because the first one was uploaded.

	rql := model.Files.T().GetAllByIndex("name", createFileRequest.Name)
	var files []schema.File
	err = model.Files.Qs(session).Rows(rql, &files)
	fmt.Println("len of files = ", len(files))
	for _, f := range files {
		fmt.Printf("file = %#v\n", f)
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
