package dir

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestFlatten(t *testing.T) {
	d, _ := Load("testdir")
	files := d.Flatten()
	if len(files) != 8 {
		t.Fatalf("Wrong number of entries in testdir - was the test data changed?")
	}

	for i, f := range files {
		switch i {
		case 0:
			checkPath("testdir", f, t)
		case 1:
			checkPath("testdir/dir1", f, t)
		case 2:
			checkPath("testdir/dir2", f, t)
		case 3:
			checkPath("testdir/dir2/dir2.f1", f, t)
		case 4:
			checkPath("testdir/dir2/dir22", f, t)
		case 5:
			checkPath("testdir/dir2/dir22/dir22.f1", f, t)
		case 6:
			checkPath("testdir/f1", f, t)
		case 7:
			checkPath("testdir/f2", f, t)
		}
	}
}

func checkPath(path string, finfo FileInfo, t *testing.T) {
	if path != finfo.Path {
		t.Fatalf("Expect path = %s, goth %s", path, finfo.Path)
	}
}
