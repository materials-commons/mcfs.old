package user

import (
	"fmt"
	"path/filepath"
	"testing"
)

var _ = fmt.Printf

func TestCreateNewUser(t *testing.T) {
	u, err := NewUserFrom("../test_data")

	if err != nil {
		t.Fatalf("NewUserFrom returned an error\n")
	}

	if u.Username == "" {
		t.Fatalf("No username\n")
	}

	if u.APIKey == "" {
		t.Fatalf("No apikey\n")
	}

	expectedPath := filepath.Join("..", "test_data", ".materials")
	if u.DotMaterialsPath() != expectedPath {
		t.Fatalf("DotMaterialsPath expected %s, got %s\n", expectedPath, u.DotMaterialsPath())
	}
}

func TestSaveUser(t *testing.T) {
	u, _ := NewUserFrom("../test_data")
	u.APIKey = "abc123"
	err := u.Save()
	if err != nil {
		t.Fatalf("Save returned error %s\n", err.Error())
	}

	u2, _ := NewUserFrom("../test_data")
	if u2.APIKey != "abc123" {
		t.Fatalf("Expected apikey to be abc123, got %s\n", u2.APIKey)
	}
}
