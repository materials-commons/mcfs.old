package access

import (
	"fmt"
	"github.com/materials-commons/base/db"
	"github.com/materials-commons/mcfs"
	"testing"
	"time"
)

var _ = fmt.Println

func TestNoServerRunning(t *testing.T) {
	_, err := GetUserByAPIKey("test")
	if err != mcfs.ErrServerNotRunning {
		t.Fatalf("Made request with no server, expected mcfs.ErrServerNotRunning, got %s", err)
	}
}

func TestWithServerRunning(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")
	var fakeStopChannel chan struct{}
	server.Init()
	go server.Run(fakeStopChannel)
	time.Sleep(1000)
	u, err := GetUserByAPIKey("test")
	if err != nil {
		t.Fatalf("Failed retrieving APIKey test: %s", err)
	}

	if u.ID != "test@mc.org" {
		t.Fatalf("Expected a different user to be retrieved: %#v", u)
	}

	u, err = GetUserByAPIKey("no-such-key")
	if err == nil {
		t.Fatalf("Sent invalid key and got a good response: %#v", u)
	}
}
