package dai

import (
	"testing"

	"github.com/materials-commons/mcfs/base/db"
)

func TestRDirsByID(t *testing.T) {
	s, err := db.RSession()
	if err != nil {
		t.Fatalf("Unable to get a session: %s", err)
	}

	rdirs := newRDirs(s)

	// Test id that doesn't exist
	if _, err := rdirs.ByID("does-not-exist"); err == nil {
		t.Fatalf("Found dir for id that doesn't exist: 'does-not-exist'")
	}
}
