package service

import (
	"fmt"
	"github.com/materials-commons/base/db"
	//"github.com/materials-commons/base/schema"
	"testing"
)

var _ = fmt.Println

func TestByID(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")
	rusers := newRUsers()

	// Test existing
	u, err := rusers.ByID("test@mc.org")
	if err != nil {
		t.Fatalf("Unable to retrieve existing user test@mc.org %s", err)
	}

	fmt.Printf("u = %#v\n", u)
}

func TestByAPIKey(t *testing.T) {

}

func TestAll(t *testing.T) {

}
