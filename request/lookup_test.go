package request

import (
	"fmt"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println

type lookupTest struct {
	field    string
	value    string
	limitTo  string
	errorNil bool
	comment  string
}

var dataDirTests = []lookupTest{
	{"id", "abc123", "", false, "No such id"},
	{"id", "73104436-5236-4adc-a0cb-00ad97286593", "", true, "id Existing with permissions"},
	{"id", "4b882c3d-0c03-4523-9ec8-f6320851a560", "", false, "id Existing without permission"},
	{"blah", "blah", "", false, "No such field"},
	{"name", "blah", "904886a7-ea57-4de7-8125-6e18c9736fd0", false, "No such name"},
	{"name", "WE43 Heat Treatments/AT 200C", "904886a7-ea57-4de7-8125-6e18c9736fd0", true, "Existing name with perimissions"},
	{"name", "WE43 Heat Treatments/AT 200C", "no-such-project", false, "Existing name with bad project"},
	{"name", "Synthetic Tooth/Presentation/1F vs Enamel/Mass Spec Compared", "34520277-4a0d-4f79-a30c-b63886f003c4", false, "Existing name without perimissions"},
}

func TestLookupDataDir(t *testing.T) {
	conductTest(t, dataDirTests, "datadir")
}

var dataFileTests = []lookupTest{
	{"id", "abc123", "", false, "No such id"},
	{"id", "1a455b46-a560-472e-acec-c96482fd655a", "e70bfd9e-9c43-4a26-b89f-c5f5ab639a72", true, "id Existing with permissions"},
	{"id", "", "", false, "id Existing without permission"},
	{"blah", "blah", "", false, "No such field"},
	{"name", "blah", "", false, "No such name"},
	{"name", "8H-4.JPG", "962d5ee5-6974-48cc-b142-a7a854374cf1", true, "Existing name with perimissions"},
	{"name", "8H-4.JPG", "blah", false, "Existing name with bad datadir"},
	{"name", "tooth-F.rrng", "ad6c499d-9309-4472-8237-a93161dbe5f1", false, "Existing name without perimissions"},
}

func TestLookupDataFile(t *testing.T) {
	conductTest(t, dataFileTests, "datafile")
}

var projectTests = []lookupTest{
	{"id", "abc123", "", false, "No such id"},
	{"id", "904886a7-ea57-4de7-8125-6e18c9736fd0", "", true, "id Existing project with permissions"},
	{"id", "34520277-4a0d-4f79-a30c-b63886f003c4", "", false, "id Existing project without permissions"},
	{"name", "WE43 Heat Treatments", "", true, "name Lookup existing"},
	{"name", "Does not exist", "", false, "name Lookup bad project name"},
	{"name", "Synthetic Tooth", "", false, "name Lookup existing but no permissions"},
}

func TestLookupProject(t *testing.T) {
	conductTest(t, projectTests, "project")
}

var invalidItemTests = []lookupTest{
	{"id", "gtarcea@umich.edu", "", false, "id Lookup users table should fail"},
}

func TestLookupInvalidItem(t *testing.T) {
	conductTest(t, invalidItemTests, "user")
}

func conductTest(t *testing.T, tests []lookupTest, whichType string) {
	h := NewReqHandler(nil, session, "")
	h.user = "gtarcea@umich.edu"
	for _, test := range tests {
		req := &protocol.LookupReq{
			Field:     test.field,
			Value:     test.value,
			LimitToID: test.limitTo,
			Type:      whichType,
		}

		v, err := h.lookup(req)
		var _ = v
		//fmt.Printf("%s/%#v/%#v\n", err, test, v)
		switch {
		case err != nil && test.errorNil:
			// Expected error to be nil
			t.Fatalf("Expected error to be nil for test %s, err %s", test.comment, err)
		case err == nil && !test.errorNil:
			// Expected error not to be nil
			t.Fatalf("Expected err != nil for test %s", test.comment)
		}
	}
}
