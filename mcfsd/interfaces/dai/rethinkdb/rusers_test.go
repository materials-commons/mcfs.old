package rethinkdb

import (
	"fmt"
	"testing"

	"github.com/materials-commons/mcfs/testutils/tdb"
)

var _ = fmt.Println

func TestRUsersByID(t *testing.T) {
	rusers := NewRUsers(tdb.NewSession())

	// Test existing
	u, err := rusers.ByID("test@mc.org")
	if err != nil {
		t.Fatalf("Unable to retrieve existing user test@mc.org %s", err)
	}

	if u.ID != "test@mc.org" {
		t.Fatalf("Wrong user retrieved expected test@mc.org, got user %#v", u)
	}

	// Test non-existant user
	u, err = rusers.ByID("does@not.exist")
	if err == nil {
		t.Fatalf("Retrieved non existant user does@not.exist")
	}
}

func TestRUsersByAPIKey(t *testing.T) {
	rusers := NewRUsers(tdb.NewSession())

	// Test existing
	u, err := rusers.ByAPIKey("test")
	if err != nil {
		t.Fatalf("Failed to retreive apikey test %s", err)
	}

	if u.ID != "test@mc.org" {
		t.Fatalf("Wrong user with apikey test: %#v", u)
	}

	// Test non-existant key
	u, err = rusers.ByAPIKey("no-such-key")
	if err == nil {
		t.Fatalf("Retrieved key that does not exist, got %#v", u)
	}
}

func TestRUsersAll(t *testing.T) {
	rusers := NewRUsers(tdb.NewSession())

	users, err := rusers.All()
	if err != nil {
		t.Fatalf("Failed retrieving all users: %s", err)
	}

	// Test that test@mc.org is one of the users
	found := false
	for _, user := range users {
		if user.ID == "test@mc.org" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("List of all users did not contain test@mc.org: %#v", users)
	}
}
