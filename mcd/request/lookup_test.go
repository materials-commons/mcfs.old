package request

import (
	"fmt"
	"github.com/materials-commons/mcfs/mcd"
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

func init() {
	mcfs.InitRethinkDB()
}

/*
Lookup datadirs. When id field we do a direct lookup. When field other than id,
then limitTo is a project_id that we will look up a directory in.
*/
var dataDirTests = []lookupTest{
	{"id", "abc123", "", false, "No such id"},
	{"id", "d0b001c6-fc0a-4e95-97c3-4427de68c0a5", "", true, "id Existing with permissions"},
	{"id", "a87806ac-8f56-4eb9-abfb-6bfbb7a19dd6", "", false, "id Existing without permission"},
	{"blah", "blah", "", false, "No such field"},
	{"name", "blah", "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", false, "No such name"},
	{"name", "Test/AT 250C/AT 2 hours", "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", true, "Existing name with perimissions"},
	{"name", "Test/AT 250C/AT 2 hours", "no-such-project", false, "Existing name with bad project"},
	{"name", "Test2/dir1", "12b5aecb-1def-463d-8886-92f6e81bf234", false, "Existing name without perimissions"},
}

func TestLookupDataDir(t *testing.T) {
	conductTest(t, dataDirTests, "datadir")
}

/*
When id lookup directly. When other field, limitTo is a datadir id.
*/
var dataFileTests = []lookupTest{
	{"id", "abc123", "", false, "No such id"},
	{"id", "692a623d-ee26-4a40-aee6-dbfa5413aefe", "", true, "id Existing with permissions"},
	{"id", "eb402860-0c6c-433b-b5b6-e0280d421461", "", false, "id Existing without permission"},
	{"blah", "blah", "", false, "No such field"},
	{"name", "blah", "", false, "No such name"},
	{"name", "R38_03085 Sample Info.txt", "c3d72271-4a32-4080-a6a3-b4c6a5c4b986", true, "Existing name with perimissions"},
	{"name", "R38_03085 Sample Info.txt", "blah", false, "Existing name with bad datadir"},
	{"name", "file1.txt", "a87806ac-8f56-4eb9-abfb-6bfbb7a19dd6", false, "Existing name without perimissions"},
}

func TestLookupDataFile(t *testing.T) {
	conductTest(t, dataFileTests, "datafile")
}

/*
When id lookup directly. When other field, limitTo is ignored.
*/
var projectTests = []lookupTest{
	{"id", "abc123", "", false, "No such id"},
	{"id", "9b18dac4-caff-4dc6-9a18-ae5c6b9c9ca3", "", true, "id Existing project with permissions"},
	{"id", "12b5aecb-1def-463d-8886-92f6e81bf234", "", false, "id Existing project without permissions"},
	{"name", "Test", "", true, "name Lookup existing with permissions"},
	{"name", "Does not exist", "", false, "name Lookup bad project name"},
	{"name", "Test2", "", false, "name Lookup existing but no permissions"},
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
	h := NewReqHandler(nil, "")
	h.user = "test@mc.org"
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
			t.Fatalf("Expected error to be nil for test type %s, test %s err %s", whichType, test.comment, err)
		case err == nil && !test.errorNil:
			// Expected error not to be nil
			t.Fatalf("Expected err != nil for test type %s, test %s", whichType, test.comment)
		}
	}
}
