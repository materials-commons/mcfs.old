package request

import (
	"fmt"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println

func TestStat(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "gtarcea@umich.edu"

	statRequest := protocol.StatReq{
		DataFileID: "1a455b46-a560-472e-acec-c96482fd655a",
	}

	resp, err := h.stat(&statRequest)

	if err != nil {
		t.Fatalf("Bad stat request %s", err)
	}

	if len(resp.DataDirs) != 1 {
		t.Fatalf("DataDirs length incorrect, expected 1 got %d", len(resp.DataDirs))
	}

	if resp.DataDirs[0] != "e70bfd9e-9c43-4a26-b89f-c5f5ab639a72" {
		t.Fatalf("Datadirs[0] incorrect = %s", resp.DataDirs[0])
	}

	if resp.Name != "R38_03085-v01_MassSpectrum.csv" {
		t.Fatalf("Name incorrect = %s", resp.Name)
	}

	if resp.Checksum != "6a600da8fe52310128ba7f193f6bb345" {
		t.Fatalf("Checksum incorrect = %s", resp.Checksum)
	}

	if resp.Size != 20637765 {
		t.Fatalf("Size incorrect = %d", resp.Size)
	}

	// Test file we don't have access to
	statRequest.DataFileID = "01cc4163-8c6f-4832-8c7b-15e34e4368ae"
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
