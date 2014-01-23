package request

import (
	"crypto/md5"
	"fmt"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/protocol"
	"io/ioutil"
	"os"
	"testing"
)

var _ = fmt.Println

func TestUploadCasesFile(t *testing.T) {
	// Test New File
	h := NewReqHandler(nil, session, "")
	h.user = "gtarcea@umich.edu"

	// Test bad upload with non existant DataFileID
	uploadReq := protocol.UploadReq{
		DataFileID: "does not exist",
		Size:       6,
		Checksum:   "abc123",
	}

	resp, err := h.upload(&uploadReq)
	if err == nil {
		t.Fatalf("Upload req succeeded with a non existant datafile id")
	}

	// Test create file and then upload
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "c33edab7-a65f-478e-9fa6-9013271c73ea",
		DataDirID: "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys_Data",
		Name:      "testfile1.txt",
		Size:      6,
		Checksum:  "abc123",
	}

	createResp, _ := h.createFile(&createFileRequest)
	createdId := createResp.ID

	uploadReq.DataFileID = createdId

	resp, err = h.upload(&uploadReq)
	if err != nil {
		t.Fatalf("Failed to start upload on a valid file %s", err)
	}

	if resp.DataFileID != createdId {
		t.Fatalf("Upload created a new version expected id %s, got %s", createdId, resp.DataFileID)
	}

	if resp.Offset != 0 {
		t.Fatalf("Upload asking for offset different than 0 (%d)", resp.Offset)
	}

	// Test create and then upload with size larger
	uploadReq.Size = 7
	resp, err = h.upload(&uploadReq)
	if err == nil {
		t.Fatalf("Upload with different size should have failed")
	}

	// Test create and then upload with size smaller
	uploadReq.Size = 5
	resp, err = h.upload(&uploadReq)
	if err == nil {
		t.Fatalf("Upload with different size should have failed")
	}

	// Test create and then upload with different checksum
	uploadReq.Size = 6
	uploadReq.Checksum = "def456"
	resp, err = h.upload(&uploadReq)
	if err == nil {
		t.Fatalf("Upload with different checksum should have failed")
	}

	// Test create and then upload with different size and checksum
	uploadReq.Size = 7
	uploadReq.Checksum = "def456"
	resp, err = h.upload(&uploadReq)
	if err == nil {
		t.Fatalf("Upload with different checksum should have failed")
	}

	// Test Existing without permissions
	h.user = "mcfada@umich.edu"
	uploadReq.Size = 6
	uploadReq.Checksum = "abc123"
	resp, err = h.upload(&uploadReq)
	if err == nil {
		t.Fatalf("Allowing upload when user doesn't have permission")
	}

	// Test interrupted transfer
	h.mcdir = "/tmp/mcdir"
	h.user = "gtarcea@umich.edu"
	os.MkdirAll(h.mcdir, 0777)
	w, err := datafileOpen(h.mcdir, createdId, 0)
	w.Write([]byte("Hello"))
	w.(*os.File).Sync()

	resp, err = h.upload(&uploadReq)
	if err != nil {
		t.Fatalf("Restart interrupted failed")
	}
	if resp.Offset != 5 {
		t.Fatalf("Offset computation incorrect")
	}
	if resp.DataFileID != createdId {
		t.Fatalf("Tried to create a new datafile id for an interrupted transfer")
	}

	// Test new version with previous interrupted
	uploadReq.Size = 8
	uploadReq.Checksum = "def456"
	resp, err = h.upload(&uploadReq)
	if err == nil {
		t.Fatalf("Allowed to create a new version when a previous version hasn't completed upload")
	}

	// Test new version when previous version has completed the upload
	w.Write([]byte("s")) // Get file to correct size to complete upload
	w.Close()
	resp, err = h.upload(&uploadReq)
	if err != nil {
		t.Fatalf("Cannot create new version of file already uploaded %s", err)
	}

	if resp.DataFileID == createdId {
		t.Fatalf("New ID was not assigned for new version of file")
	}

	if resp.Offset != 0 {
		t.Fatalf("Uploading new version offset should be 0 not %d", resp.Offset)
	}

	fmt.Println("Deleting datafile id", createdId)
	model.Delete("datafiles", createdId, session)
	os.RemoveAll("/tmp/mcdir")
}

func TestUploadNewFile(t *testing.T) {
	h := NewReqHandler(nil, session, "/tmp/mcdir")
	h.user = "gtarcea@umich.edu"
	testfilePath := "/tmp/mcdir/testfile.txt"
	testfileData := "Hello world for testing"
	testfileLen := int64(len(testfileData))

	// Create file that we are going to upload
	os.MkdirAll("/tmp/mcdir", 0777)
	ioutil.WriteFile(testfilePath, []byte(testfileData), 0777)
	checksum, _ := file.Hash(md5.New(), testfilePath)
	checksumHex := fmt.Sprintf("%x", checksum)
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "c33edab7-a65f-478e-9fa6-9013271c73ea",
		DataDirID: "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys_Data",
		Name:      "testfile.txt",
		Size:      testfileLen,
		Checksum:  checksumHex,
	}

	createResp, _ := h.createFile(&createFileRequest)
	createdId := createResp.ID
	defer cleanup(createdId)
	uploadReq := protocol.UploadReq{
		DataFileID: createdId,
		Size:       testfileLen,
		Checksum:   checksumHex,
	}

	resp, err := h.upload(&uploadReq)
	if err != nil {
		t.Fatalf("error %s", err)
	}

	if resp.DataFileID != createdId {
		t.Fatalf("ids don't match")
	}

	if resp.Offset != 0 {
		t.Fatalf("Wrong offset")
	}

	uploadHandler, err := prepareUploadHandler(h, resp.DataFileID, resp.Offset)
	if err != nil {
		t.Fatalf("Couldn't create uploadHandler %s", err)
	}

	var _ = uploadHandler
	sendReq := protocol.SendReq{
		DataFileID: createdId,
		Bytes:      []byte(testfileData),
	}

	n, err := uploadHandler.sendReqWrite(&sendReq)
	if n != len(testfileData) {
		t.Fatalf("Incorrect number of bytes written expected %d, wrote %d", testfileLen, n)
	}

	nchecksum, err := file.Hash(md5.New(), datafilePath(h.mcdir, createdId))
	if err != nil {
		t.Fatalf("Unable to checksum datafile %s", createdId)
	}

	dfClose(uploadHandler.w, uploadHandler.dataFileID, uploadHandler.session)

	nchecksumHex := fmt.Sprintf("%x", nchecksum)
	if nchecksumHex != checksumHex {
		t.Fatalf("Checksums don't match for uploaded file expected = %s, got %s", checksumHex, nchecksumHex)
	}
}

func TestPartialToCompleted(t *testing.T) {
	h := NewReqHandler(nil, session, "/tmp/mcdir")
	h.user = "gtarcea@umich.edu"
	testfilePath := "/tmp/mcdir/testfile.txt"
	testfileData := "Hello world for testing"
	testfileLen := int64(len(testfileData))

	// Create file that we are going to upload
	os.MkdirAll("/tmp/mcdir", 0777)
	ioutil.WriteFile(testfilePath, []byte(testfileData), 0777)
	checksum, _ := file.Hash(md5.New(), testfilePath)
	checksumHex := fmt.Sprintf("%x", checksum)
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "c33edab7-a65f-478e-9fa6-9013271c73ea",
		DataDirID: "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys_Data",
		Name:      "testfile.txt",
		Size:      testfileLen,
		Checksum:  checksumHex,
	}

	createResp, _ := h.createFile(&createFileRequest)
	createdId := createResp.ID
	defer cleanup(createdId)

	uploadReq := protocol.UploadReq{
		DataFileID: createdId,
		Size:       testfileLen,
		Checksum:   checksumHex,
	}

	resp, err := h.upload(&uploadReq)
	if err != nil {
		t.Fatalf("error %s", err)
	}

	if resp.DataFileID != createdId {
		t.Fatalf("ids don't match")
	}

	if resp.Offset != 0 {
		t.Fatalf("Wrong offset")
	}

	uploadHandler, err := prepareUploadHandler(h, resp.DataFileID, resp.Offset)
	if err != nil {
		t.Fatalf("Couldn't create uploadHandler %s", err)
	}

	sendReq := protocol.SendReq{
		DataFileID: createdId,
		Bytes:      []byte(testfileData[0:3]),
	}

	n, _ := uploadHandler.sendReqWrite(&sendReq)
	if n != 3 {
		t.Fatalf("Wrong number of bytes written, expected 3, got %d", n)
	}
	dfClose(uploadHandler.w, uploadHandler.dataFileID, uploadHandler.session)

	// Start a new uploadReq so we can finish the upload
	resp, err = h.upload(&uploadReq)
	if err != nil {
		t.Fatalf("Completing upload rejected %s", err)
	}

	if resp.DataFileID != createdId {
		t.Fatalf("Unexpected creation of a new version of datafile")
	}

	if resp.Offset != 3 {
		t.Fatalf("Wrong offset expected 3, got %d", resp.Offset)
	}

	uploadHandler, err = prepareUploadHandler(h, resp.DataFileID, resp.Offset)
	if err != nil {
		t.Fatalf("Couldn't create uploadHandler %s", err)
	}

	sendReq.Bytes = []byte(testfileData[resp.Offset:])
	n, _ = uploadHandler.sendReqWrite(&sendReq)
	if n != len(testfileData[resp.Offset:]) {
		t.Fatalf("Incorrect number of bytes written expected %d, wrote %d", testfileLen, n)
	}

	nchecksum, err := file.Hash(md5.New(), datafilePath(h.mcdir, createdId))
	if err != nil {
		t.Fatalf("Unable to checksum datafile %s", createdId)
	}

	dfClose(uploadHandler.w, uploadHandler.dataFileID, uploadHandler.session)

	nchecksumHex := fmt.Sprintf("%x", nchecksum)
	if nchecksumHex != checksumHex {
		t.Fatalf("Checksums don't match for uploaded file expected = %s, got %s", checksumHex, nchecksumHex)
	}
}

func cleanup(datafileId string) {
	fmt.Println("Deleting datafile id =", datafileId)
	model.Delete("datafiles", datafileId, session)
	os.RemoveAll("/tmp/mcdir")
}
