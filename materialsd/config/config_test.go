package config

import (
	"encoding/json"
	"fmt"
	"github.com/materials-commons/mcfs/client/user"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var _ = fmt.Printf

func TestNoConfigNoEnv(t *testing.T) {
	u, _ := user.NewUserFrom("test_data/noconfig")
	ConfigInitialize(u)
	if Config.MaterialsCommons.API != "https://api.materialscommons.org" {
		t.Fatalf("api value incorrect %s\n", Config.MaterialsCommons.API)
	}

	if Config.MaterialsCommons.URL != "https://materialscommons.org" {
		t.Fatalf("api value incorrect %s\n", Config.MaterialsCommons.URL)
	}

	if Config.MaterialsCommons.Download != "https://download.materialscommons.org" {
		t.Fatalf("api value incorrect %s\n", Config.MaterialsCommons.Download)
	}

	if Config.User.DefaultProject != "" {
		t.Fatalf("defaultProject incorrect %s\n", Config.User.DefaultProject)
	}

	expectedWebdir := filepath.Join(u.DotMaterialsPath(), "website")
	if Config.Server.Webdir != expectedWebdir {
		t.Fatalf("webdir incorrect %s, expected %s\n", Config.Server.Webdir, expectedWebdir)
	}

	if Config.Server.Port != 8081 {
		t.Fatalf("port incorrect %d\n", Config.Server.Port)
	}

	if Config.Server.SocketIOPort != 8082 {
		t.Fatalf("socket port incorrect %d\n", Config.Server.SocketIOPort)
	}

	if Config.Server.Address != "localhost" {
		t.Fatalf("address incorrect %s\n", Config.Server.Address)
	}

	if Config.Server.UpdateCheckInterval != 4*time.Hour {
		t.Fatalf("address incorrect %d\n", Config.Server.UpdateCheckInterval)
	}
}

func TestWithEnvSetting(t *testing.T) {
	u, _ := user.NewUserFrom("test_data/noconfig")
	os.Setenv("MCURL", "http://localhost")
	ConfigInitialize(u)
	if Config.MaterialsCommons.URL != "http://localhost" {
		t.Fatalf("url expected http://localhost, got %s\n", Config.MaterialsCommons.URL)
	}
}

func TestJson(t *testing.T) {
	u, _ := user.NewUserFrom("test_data/noconfig")
	ConfigInitialize(u)
	_, err := json.MarshalIndent(Config, "", "   ")
	if err != nil {
		t.Fatal(err)
	}
}
