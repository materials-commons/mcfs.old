package mcfs

import (
	"crypto/md5"
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/contrib/model"
	"github.com/materials-commons/gohandy/handyfile"
	"github.com/materials-commons/gohandy/marshaling"
	"github.com/materials-commons/materials/util"
	"github.com/materials-commons/mcfs/request"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var _ = fmt.Println

var session *r.Session
var m = util.NewChannelMarshaler()
var c = &Client{
	MarshalUnmarshaler: m,
}

const MCDir = "/tmp/mcdir"

func init() {
	os.RemoveAll(MCDir)
	session, _ = r.Connect(map[string]interface{}{
		"address":  "localhost:30815",
		"database": "materialscommons",
	})
	go mcfsServer(m)
}

func TestLoginLogout(t *testing.T) {

	err := c.Login("gtarcea@umich.edu", "abc123")
	if err == nil {
		t.Fatalf("Login accepted with bad key")
	}

	err = c.Login("gtarcea@umich.edu", "472abe203cd411e3a280ac162d80f1bf")
	if err != nil {
		t.Fatalf("Login should have succeeded %s", err)
	}

	err = c.Logout()
	if err != nil {
		t.Fatalf("Logout failed %s", err)
	}
}

func TestUploadNewFile(t *testing.T) {
	c.Login("gtarcea@umich.edu", "472abe203cd411e3a280ac162d80f1bf")
	fileData := "Hello world from Materials Commons"
	filePath := filepath.Join(MCDir, "testnewfile.txt")
	ioutil.WriteFile(filePath, []byte(fileData), 0777)
	projectID := "c33edab7-a65f-478e-9fa6-9013271c73ea"
	dataDirID := "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys_Data"
	uploaded, dataFileID, err := c.UploadNewFile(projectID, dataDirID, filePath)
	fmt.Printf("%d/%s/%s\n", uploaded, dataFileID, err)
	if err != nil {
		t.Fatalf("Upload unexpectedly failed %s", err)
	}

	if int64(len(fileData)) != uploaded {
		t.Fatalf("Upload count (%d) different than size of data (%d)", uploaded, len(fileData))
	}
	dataFilePath := request.DataFilePath(MCDir, dataFileID)
	dataFileChecksum, dataFileSize, err := fileInfo(dataFilePath)

	if err != nil {
		t.Fatalf("Failed to checksum datafile %s", dataFilePath)
	}

	fileChecksum, fileSize, err := fileInfo(filePath)
	if err != nil {
		t.Fatalf("Failed to checksum file %s", filePath)
	}

	if fileSize != dataFileSize {
		t.Fatalf("File sizes did not match %d/%d", dataFileSize, fileSize)
	}

	if dataFileChecksum != fileChecksum {
		t.Fatalf("Checksums did not match %s/%s", dataFileChecksum, fileChecksum)
	}

	defer cleanup(dataFileID)
}

func TestRestartFileUpload(t *testing.T) {
	fileData := "Hello world from Materials Commons"
	filePath := filepath.Join(MCDir, "testnewfilerestart.txt")
	ioutil.WriteFile(filePath, []byte(fileData), 0777)
	filePathPartial := filepath.Join(MCDir, "testnewfilerestartpartial.txt")
	ioutil.WriteFile(filePathPartial, []byte(fileData[:10]), 0777)
	realChecksum, err := handyfile.HashStr(md5.New(), filePath)
	var _ = realChecksum
	realSize := len(fileData)
	var _ = realSize
	projectID := "c33edab7-a65f-478e-9fa6-9013271c73ea"
	dataDirID := "gtarcea@umich.edu$Test_Proj_6111_Aluminum_Alloys_Data"
	uploaded, dataFileID, err := c.UploadNewFile(projectID, dataDirID, filePathPartial)

	if err != nil {
		t.Fatalf("Failed to upload partial data %s", err)
	}

	if uploaded != 10 {
		t.Fatalf("Wrong number of bytes written expected %d, got %d", 10, uploaded)
	}

	// We have uploaded a partial of the file. Too fool the system we now need
	// to update the database with the real size, checksum and name. Then we
	// can "restart" the download.
	r.Table("datafiles").Get(dataFileID).Update(map[string]interface{}{
		"checksum": realChecksum,
		"size":     realSize,
		"name":     "testnewfilerestart.txt",
	}).RunWrite(session)

	n, err := c.RestartFileUpload(dataFileID, filePath)

	if err != nil {
		t.Fatalf("Failed to restart upload %s", err)
	}

	if n != int64(len(fileData)-10) {
		t.Fatalf("Wrong number of bytes written expected %d, got %d", len(fileData)-10, n)
	}

	dataFilePath := request.DataFilePath(MCDir, dataFileID)
	dataFileChecksum, dataFileSize, err := fileInfo(dataFilePath)

	if err != nil {
		t.Fatalf("Failed to checksum datafile %s", dataFilePath)
	}

	fileChecksum, fileSize, err := fileInfo(filePath)
	if err != nil {
		t.Fatalf("Failed to checksum file %s", filePath)
	}

	if fileSize != dataFileSize {
		t.Fatalf("File sizes did not match %d/%d", dataFileSize, fileSize)
	}

	if dataFileChecksum != fileChecksum {
		t.Fatalf("Checksums did not match %s/%s", dataFileChecksum, fileChecksum)
	}
	defer cleanup(dataFileID)
}

func cleanup(dataFileID string) {
	fmt.Println("Deleting datafile id =", dataFileID)
	model.Delete("datafiles", dataFileID, session)
}

func mcfsServer(m marshaling.MarshalUnmarshaler) {
	h := request.NewReqHandler(m, session, MCDir)
	os.MkdirAll("/tmp/mcdir", 0777)
	h.Run()
}
