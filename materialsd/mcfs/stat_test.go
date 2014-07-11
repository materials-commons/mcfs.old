package mcfs

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestStatProject(t *testing.T) {
	ps, err := c.StatProject("Test")

	if err != nil {
		t.Fatalf("Failed to stat project 'test': %s", err)
	}

	if len(ps.Entries) == 0 {
		t.Fatalf("Stat failed, no entries")
	}

	for _, entry := range ps.Entries {
		fmt.Printf("%#v\n", entry)
	}
}
