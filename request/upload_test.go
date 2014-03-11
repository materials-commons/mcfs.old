package request

import (
	"crypto/md5"
	"fmt"
	"github.com/materials-commons/base/model"
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
	h.user = "test@mc.org"

	// Test bad upload with non existant DataFileID
	uploadReq := protocol.UploadReq{
		DataFileID: "does not exist",
		Size:       6,
		Checksum:   "abc123",
	}

	resp, status := h.upload(&uploadReq)
	if status == nil {
		t.Fatalf("Upload req succeeded with a non existant datafile id")
	}

	// Test create file and then upload
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "testfile1.txt",
		Size:      6,
		Checksum:  "abc123",
	}

	createResp, _ := h.createFile(&createFileRequest)
	createdID := createResp.ID

	uploadReq.DataFileID = createdID

	resp, status = h.upload(&uploadReq)
	if status != nil {
		t.Fatalf("Failed to start upload on a valid file %s", status.err)
	}

	if resp.DataFileID != createdID {
		t.Fatalf("Upload created a new version expected id %s, got %s", createdID, resp.DataFileID)
	}

	if resp.Offset != 0 {
		t.Fatalf("Upload asking for offset different than 0 (%d)", resp.Offset)
	}

	// Test create and then upload with size larger
	uploadReq.Size = 7
	resp, status = h.upload(&uploadReq)
	if status == nil {
		t.Fatalf("Upload with different size should have failed")
	}

	// Test create and then upload with size smaller
	uploadReq.Size = 5
	resp, status = h.upload(&uploadReq)
	if status == nil {
		t.Fatalf("Upload with different size should have failed")
	}

	// Test create and then upload with different checksum
	uploadReq.Size = 6
	uploadReq.Checksum = "def456"
	resp, status = h.upload(&uploadReq)
	if status == nil {
		t.Fatalf("Upload with different checksum should have failed")
	}

	// Test create and then upload with different size and checksum
	uploadReq.Size = 7
	uploadReq.Checksum = "def456"
	resp, status = h.upload(&uploadReq)
	if status == nil {
		t.Fatalf("Upload with different checksum should have failed")
	}

	// Test Existing without permissions
	h.user = "test2@mc.org"
	uploadReq.Size = 6
	uploadReq.Checksum = "abc123"
	resp, status = h.upload(&uploadReq)
	if status == nil {
		t.Fatalf("Allowing upload when user doesn't have permission")
	}

	// Test interrupted transfer
	h.mcdir = "/tmp/mcdir"
	h.user = "test@mc.org"
	os.MkdirAll(h.mcdir, 0777)
	w, _ := datafileOpen(h.mcdir, createdID, 0)
	w.Write([]byte("Hello"))
	w.(*os.File).Sync()

	resp, status = h.upload(&uploadReq)
	if status != nil {
		t.Fatalf("Restart interrupted failed")
	}
	if resp.Offset != 5 {
		t.Fatalf("Offset computation incorrect")
	}
	if resp.DataFileID != createdID {
		t.Fatalf("Tried to create a new datafile id for an interrupted transfer")
	}

	// Test new version with previous interrupted
	uploadReq.Size = 8
	uploadReq.Checksum = "def456"
	resp, status = h.upload(&uploadReq)
	if status == nil {
		t.Fatalf("Allowed to create a new version when a previous version hasn't completed upload")
	}

	// Test new version when previous version has completed the upload
	w.Write([]byte("s")) // Get file to correct size to complete upload
	w.Close()
	resp, status = h.upload(&uploadReq)
	if status != nil {
		t.Fatalf("Cannot create new version of file already uploaded %s", status.err)
	}

	if resp.DataFileID == createdID {
		t.Fatalf("New ID was not assigned for new version of file")
	}

	if resp.Offset != 0 {
		t.Fatalf("Uploading new version offset should be 0 not %d", resp.Offset)
	}

	fmt.Println("Deleting datafile id", createdID)
	model.Delete("datafiles", createdID, session)
	os.RemoveAll("/tmp/mcdir")
}

func TestUploadNewFile(t *testing.T) {
	h := NewReqHandler(nil, session, "/tmp/mcdir")
	h.user = "test@mc.org"
	testfilePath := "/tmp/mcdir/testfile.txt"
	testfileData := "Hello world for testing"
	testfileLen := int64(len(testfileData))

	// Create file that we are going to upload
	os.MkdirAll("/tmp/mcdir", 0777)
	ioutil.WriteFile(testfilePath, []byte(testfileData), 0777)
	checksum, _ := file.Hash(md5.New(), testfilePath)
	checksumHex := fmt.Sprintf("%x", checksum)
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "testfile.txt",
		Size:      testfileLen,
		Checksum:  checksumHex,
	}

	createResp, _ := h.createFile(&createFileRequest)
	createdID := createResp.ID
	defer cleanup(createdID)
	uploadReq := protocol.UploadReq{
		DataFileID: createdID,
		Size:       testfileLen,
		Checksum:   checksumHex,
	}

	resp, status := h.upload(&uploadReq)
	if status != nil {
		t.Fatalf("error %s", status.err)
	}

	if resp.DataFileID != createdID {
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
		DataFileID: createdID,
		Bytes:      []byte(testfileData),
	}

	n, status := uploadHandler.sendReqWrite(&sendReq)
	if n != len(testfileData) {
		t.Fatalf("Incorrect number of bytes written expected %d, wrote %d", testfileLen, n)
	}

	nchecksum, err := file.Hash(md5.New(), datafilePath(h.mcdir, createdID))
	if err != nil {
		t.Fatalf("Unable to checksum datafile %s", createdID)
	}

	dfClose(uploadHandler.w, uploadHandler.dataFileID, uploadHandler.session)

	nchecksumHex := fmt.Sprintf("%x", nchecksum)
	if nchecksumHex != checksumHex {
		t.Fatalf("Checksums don't match for uploaded file expected = %s, got %s", checksumHex, nchecksumHex)
	}
}

func TestPartialToCompleted(t *testing.T) {
	h := NewReqHandler(nil, session, "/tmp/mcdir")
	h.user = "test@mc.org"
	testfilePath := "/tmp/mcdir/testfile.txt"
	testfileData := "Hello world for testing"
	testfileLen := int64(len(testfileData))

	// Create file that we are going to upload
	os.MkdirAll("/tmp/mcdir", 0777)
	ioutil.WriteFile(testfilePath, []byte(testfileData), 0777)
	checksum, _ := file.Hash(md5.New(), testfilePath)
	checksumHex := fmt.Sprintf("%x", checksum)
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "testfile.txt",
		Size:      testfileLen,
		Checksum:  checksumHex,
	}

	createResp, _ := h.createFile(&createFileRequest)
	createdID := createResp.ID
	defer cleanup(createdID)

	uploadReq := protocol.UploadReq{
		DataFileID: createdID,
		Size:       testfileLen,
		Checksum:   checksumHex,
	}

	resp, status := h.upload(&uploadReq)
	if status != nil {
		t.Fatalf("error %s", status.err)
	}

	if resp.DataFileID != createdID {
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
		DataFileID: createdID,
		Bytes:      []byte(testfileData[0:3]),
	}

	n, _ := uploadHandler.sendReqWrite(&sendReq)
	if n != 3 {
		t.Fatalf("Wrong number of bytes written, expected 3, got %d", n)
	}
	dfClose(uploadHandler.w, uploadHandler.dataFileID, uploadHandler.session)

	// Start a new uploadReq so we can finish the upload
	resp, status = h.upload(&uploadReq)
	if status != nil {
		t.Fatalf("Completing upload rejected %s", status.err)
	}

	if resp.DataFileID != createdID {
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

	nchecksum, err := file.Hash(md5.New(), datafilePath(h.mcdir, createdID))
	if err != nil {
		t.Fatalf("Unable to checksum datafile %s", createdID)
	}

	dfClose(uploadHandler.w, uploadHandler.dataFileID, uploadHandler.session)

	nchecksumHex := fmt.Sprintf("%x", nchecksum)
	if nchecksumHex != checksumHex {
		t.Fatalf("Checksums don't match for uploaded file expected = %s, got %s", checksumHex, nchecksumHex)
	}
}

func TestUploadNewFileExistingFileMatches(t *testing.T) {
	h := NewReqHandler(nil, session, "/tmp/mcdir")
	h.user = "test@mc.org"
	testfilePath := "/tmp/mcdir/testfile.txt"
	testfileData := "Hello world for testing"
	testfileLen := int64(len(testfileData))

	// Create file that we are going to upload
	os.MkdirAll("/tmp/mcdir", 0777)
	ioutil.WriteFile(testfilePath, []byte(testfileData), 0777)
	checksum, _ := file.Hash(md5.New(), testfilePath)
	checksumHex := fmt.Sprintf("%x", checksum)
	createFileRequest := protocol.CreateFileReq{
		ProjectID: "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3",
		DataDirID: "f0ebb733-c75d-4983-8d68-242d688fcf73",
		Name:      "testfile.txt",
		Size:      testfileLen,
		Checksum:  checksumHex,
	}

	createResp, _ := h.createFile(&createFileRequest)
	createdID := createResp.ID
	defer cleanup(createdID)
	uploadReq := protocol.UploadReq{
		DataFileID: createdID,
		Size:       testfileLen,
		Checksum:   checksumHex,
	}

	resp, status := h.upload(&uploadReq)
	if status != nil {
		t.Fatalf("error %s", status.err)
	}

	uploadHandler, err := prepareUploadHandler(h, resp.DataFileID, resp.Offset)
	if err != nil {
		t.Fatalf("Couldn't create uploadHandler %s", err)
	}

	sendReq := protocol.SendReq{
		DataFileID: createdID,
		Bytes:      []byte(testfileData[:len(testfileData)-1]),
	}

	n, _ := uploadHandler.sendReqWrite(&sendReq)
	if n != len(testfileData)-1 {
		t.Fatalf("Wrong number of bytes written, expected %d, got %d", testfileLen, n)
	}
	dfClose(uploadHandler.w, uploadHandler.dataFileID, uploadHandler.session)

	// Now we are going to try and upload the same file to a different
	// datadir. The system should detect that we have already uploaded
	// the file and send us back the id from the file created above.
	//
	// There are two cases. Above we only wrote a partial file for the original file, so we should
	// get back the origin file id and an offset, even though a new id was created from the
	// second create file call.
	//
	createFileRequest.DataDirID = "c3d72271-4a32-4080-a6a3-b4c6a5c4b986"
	createResp, status = h.createFile(&createFileRequest)
	if status != nil {
		t.Errorf("Failed to create new file: %s", status.err)
	}

	newID := createResp.ID
	uploadReq.DataFileID = newID
	resp, status = h.upload(&uploadReq)
	if resp.DataFileID != createdID {
		t.Errorf("Wronge datafile id sent when uploading a file that matches on the server. Expected %s, got %s", createdID, resp.DataFileID)
	}

	if resp.Offset != testfileLen-1 {
		t.Errorf("Got back wrong length, got %d, expected %d", resp.Offset, testfileLen-1)
	}

	// Then we will write the rest of the file and request an upload. Now we should get back
	// the newly created id and an offset equal to the length of the file
	w, err := datafileOpen(h.mcdir, createdID, testfileLen-1)
	w.Write([]byte(testfileData[len(testfileData)-1:]))
	w.Close()
	resp, status = h.upload(&uploadReq)
	if resp.DataFileID != newID {
		t.Errorf("Wrong datafile id sent when uploading a file that matches on the server. Expected %s, got %s", newID, resp.DataFileID)
	}

	if resp.Offset != testfileLen {
		t.Errorf("Got back wrong length, got %d, expected %d", resp.Offset, testfileLen-1)
	}

	model.Delete("datafiles", newID, session)
}

func cleanup(datafileID string) {
	fmt.Println("Deleting datafile id =", datafileID)
	model.Delete("datafiles", datafileID, session)
	os.RemoveAll("/tmp/mcdir")
}
