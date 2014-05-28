package service

import (
	"testing"
)

func TestNewUsersRethinkDB(t *testing.T) {
	// Test RethinkDB
	u := NewUsers(RethinkDB)
	if _, ok := u.(rUsers); !ok {
		t.Fatalf("Requested RethinkDB interface to users and got %T instead", u)
	}
}

func TestNewUsersSQL(t *testing.T) {
	defer func() {
		s := recover()
		if s == nil {
			t.Fatalf("Did not panic for SQL")
		}
	}()

	NewUsers(SQL)
}

func TestNewUsersInvalid(t *testing.T) {
	defer func() {
		s := recover()
		if s == nil {
			t.Fatalf("Did not panic for invalid")
		}
	}()

	NewUsers(-1)
}
