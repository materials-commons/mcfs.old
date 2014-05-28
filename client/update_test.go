package materials

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Keep compiler from aborting due to an used fmt package
var _ = fmt.Printf

func TestBinaryUrl(t *testing.T) {
	oss := []string{"windows", "darwin", "linux"}
	expected := map[string]string{
		"windows": "http://localhost/windows/materials.exe",
		"darwin":  "http://localhost/darwin/materials",
		"linux":   "http://localhost/linux/materials",
	}
	url := "http://localhost"
	for _, os := range oss {
		binaryURL := binaryURLForRuntime(url, os)
		expectedURL, _ := expected[os]
		if binaryURL != expectedURL {
			t.Fatalf("Bad url %s, expected %s\n", binaryURL, expectedURL)
		}
	}

	expectedURL, _ := expected[runtime.GOOS]
	binaryURL := binaryURL(url)
	if binaryURL != expectedURL {
		t.Fatalf("Bad url %s, expected %s\n", binaryURL, expectedURL)
	}
}

func TestDownloadNewBinary(t *testing.T) {
	testBinary := map[string]string{
		"windows": "materials.test.exe",
		"darwin":  "materials.test",
		"linux":   "materials.test",
	}
	ts := httptest.NewServer(http.FileServer(http.Dir("test_data")))
	defer ts.Close()

	path, err := downloadNewBinary(binaryURLForRuntime(ts.URL, "linux"))
	if err != nil {
		t.Fatalf("Unexpected error on download %s\n", err.Error())
	}

	testBinaryName, ok := testBinary[runtime.GOOS]
	if !ok {
		panic(fmt.Sprintf("Unknown OS for test %s", runtime.GOOS))
	}

	expectedPath := filepath.Join(os.TempDir(), testBinaryName)
	if path != expectedPath {
		t.Fatalf("Downloaded to unexpected name %s, expected %s\n", path, expectedPath)
	}

	// Update this sum if you change the file test_data/linux/materials
	// Computed this by doing:
	// dlChecksum := checksumFor(path)
	// fmt.Printf("checksum = %d", dlChecksum)
	//expectedChecksum := uint32(1134331119)
	//dlChecksum := checksumFor(path)
	//if dlChecksum != expectedChecksum {
	//	t.Fatalf("Checksums don't match got: %d, expected: %d\n", dlChecksum, expectedChecksum)
	//}
}
