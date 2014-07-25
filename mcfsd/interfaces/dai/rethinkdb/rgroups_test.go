package rethinkdb

import (
	"fmt"
	"testing"

	"github.com/materials-commons/mcfs/common/schema"
	"github.com/materials-commons/mcfs/testutils/tdb"
)

var _ = fmt.Println

func TestHasAccess(t *testing.T) {
	rgroups := NewRGroups(tdb.NewSession())
	user := "gtarcea@umich.edu"
	owner := "mcfada@umich.edu"
	// Test empty table different user
	if rgroups.HasAccess(owner, "someuser@umich.edu") {
		t.Fatalf("Access passed should have failed with empty usergroups table")
	}

	//Test empty table same user
	if !rgroups.HasAccess("gtarcea@umich.edu", "gtarcea@umich.edu") {
		t.Fatalf("Access failed when user is also the user")
	}

	ug := schema.NewGroup("mcfada@umich.edu", "tgroup1")
	ug.Users = append(ug.Users, "gtarcea@umich.edu")
	g, err := rgroups.Insert(&ug)
	if err != nil {
		t.Fatalf("Unable to create new group: %s", err)
	}
	defer deleteItem(g.ID)

	// Test user who should have access
	if !rgroups.HasAccess(owner, user) {
		t.Fatalf("gtarcea@umich.edu should have had access")
	}

	// Test user who doesn't have access
	if rgroups.HasAccess(owner, "nouser@umich.edu") {
		t.Fatalf("nouser@umich.edu should not have access")
	}
}

func deleteItem(id string) {
	fmt.Printf("Deleting group id %s\n", id)
	rgroups := NewRGroups(tdb.NewSession())
	rgroups.Delete(id)
}
