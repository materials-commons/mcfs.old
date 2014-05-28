package request

import (
	"fmt"
	"github.com/materials-commons/mcfs/server/protocol"
	"testing"
)

var _ = fmt.Println

func TestStat(t *testing.T) {
	h := NewReqHandler(nil, "")
	h.user = "test@mc.org"

	statRequest := protocol.StatReq{
		DataFileID: "692a623d-ee26-4a40-aee6-dbfa5413aefe",
	}

	resp, err := h.stat(&statRequest)

	if err != nil {
		t.Fatalf("Bad stat request %s", err)
	}

	if len(resp.DataDirs) != 1 {
		t.Fatalf("DataDirs length incorrect, expected 1 got %d", len(resp.DataDirs))
	}

	if resp.DataDirs[0] != "c3d72271-4a32-4080-a6a3-b4c6a5c4b986" {
		t.Fatalf("Datadirs[0] incorrect = %s", resp.DataDirs[0])
	}

	if resp.Name != "R38_03085 Sample Info.txt" {
		t.Fatalf("Name incorrect = %s", resp.Name)
	}

	if resp.Checksum != "72d47a675e81cf4a283aaf67587ddd28" {
		t.Fatalf("Checksum incorrect = %s", resp.Checksum)
	}

	if resp.Size != 585 {
		t.Fatalf("Size incorrect = %d", resp.Size)
	}

	// Test file we don't have access to
	statRequest.DataFileID = "eb402860-0c6c-433b-b5b6-e0280d421461"
	resp, err = h.stat(&statRequest)

	if err == nil {
		t.Fatalf("Access to file we shouldn't have access to")
	}

	// Test sending bad DataFileID
	statRequest.DataFileID = "idonotexist"
	resp, err = h.stat(&statRequest)
	if err == nil {
		t.Fatalf("Succeeded for data file that doesn't exist")
	}
}
