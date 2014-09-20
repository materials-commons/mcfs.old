package materials

import (
	"encoding/json"
	"fmt"
	"github.com/materials-commons/mcfs/materialsd/config"
	"github.com/materials-commons/mcfs/materialsd/user"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

const expectedNumber = 3
const testData = "test_data/.materials/projectdb"
const corruptedData = "test_data/corrupted/.materials/projectdb"

func TestWrite(t *testing.T) {
	u, _ := user.NewUserFrom("test_data")
	config.ConfigInitialize(u)
	if true {
		return
	}
	var changes = map[string]ProjectFileChange{
		"hash1": ProjectFileChange{
			Path: "/tmp/proj1/a.txt",
			Type: "create",
			When: time.Now(),
		},
		"hash2": {
			Path: "/tmp/proj1/b.txt",
			Type: "modify",
			When: time.Now(),
		},
	}
	p := Project{
		Name:    "proj1",
		Path:    "/tmp/proj1",
		Status:  "Loaded",
		ModTime: time.Now(),
		MCId:    "abc123",
		Changes: changes,
		Ignore:  []string{"._DotFiles_.", "*.save"},
	}

	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		t.Fatalf(err.Error())
	}
	ioutil.WriteFile("/tmp/proj1.project", b, os.ModePerm)
}

func TestProjectsFrom(t *testing.T) {
	projects, err := OpenProjectDB(testData)
	if err != nil {
		t.Fatalf("TestProjectsFrom failed loading the test_data projects, %s\n", err.Error())
	}

	if len(projects.Projects()) != expectedNumber {
		t.Fatalf("Number of projects incorrect, it should have been %d: %d\n",
			expectedNumber, len(projects.Projects()))
	}
}

func TestProjectsFromWithBadDirectory(t *testing.T) {
	projects, err := OpenProjectDB("no-such-directory")
	if err == nil {
		t.Fatalf("ProjectFrom should have returned an error\n")
	}

	if projects != nil {
		t.Fatalf("projects should have been nil\n")
	}
}

func TestProjectAddDuplicate(t *testing.T) {
	p, _ := OpenProjectDB(testData)
	err := p.Add(Project{Name: "proj1", Path: "/tmp", Status: "Unloaded"})
	if err == nil {
		t.Fatalf("Duplicate project was added\n")
	}

	p2, _ := OpenProjectDB(testData)
	l := len(p2.Projects())
	if l != expectedNumber {
		for _, p := range p2.Projects() {
			fmt.Println(p)
		}
		t.Fatalf("Expected %d projects, got %d\n", expectedNumber, l)
	}
}

func TestProjectAdd(t *testing.T) {
	p, _ := OpenProjectDB(testData)
	err := p.Add(Project{Name: "new proj", Path: "/tmp", Status: "Unloaded"})
	if err != nil {
		t.Fatalf("Add failed to add new project '%s'\n", err.Error())
	}

	l := len(p.Projects())
	if l != expectedNumber+1 {
		t.Fatalf("Expected number of projects to be %d, got %d\n", expectedNumber+1, l)
	}

	p2, _ := OpenProjectDB(testData)
	l = len(p2.Projects())
	if l != expectedNumber+1 {
		t.Fatalf("Expected number of projects to be %d, got %d\n", expectedNumber+1, l)
	}
}

func TestProjectRemove(t *testing.T) {
	p, _ := OpenProjectDB(testData)
	err := p.Remove("new proj")
	if err != nil {
		t.Fatalf("Remove failed to add new project\n")
	}

	p2, _ := OpenProjectDB(testData)
	l := len(p2.Projects())
	if l != expectedNumber {
		t.Fatalf("Expected number of projects to be %d, got %d\n", expectedNumber, l)
	}
}

func TestProjectExists(t *testing.T) {
	p, _ := OpenProjectDB(testData)
	if p.Exists("does-not-exist") {
		t.Fatalf("Found project that doesn't exist\n")
	}

	if !p.Exists("proj1") {
		t.Fatalf("Failed to find project that should exist: proj1\n")
	}
}

func TestProjectFind(t *testing.T) {
	p, _ := OpenProjectDB(testData)

	_, found := p.Find("does-not-exist")
	if found {
		t.Fatalf("Found project that does not exist")
	}

	_, found = p.Find("proj1")
	if !found {
		t.Fatalf("Did not find project proj1\n")
	}

	p.Add(Project{Name: "newproj", Path: "/tmp/newproj"})
	_, found = p.Find("newproj")
	if !found {
		t.Fatalf("Did not find added project newproj\n")
	}

	p.Remove("newproj")
	_, found = p.Find("newproj")
	if found {
		t.Fatalf("Found project that was just removed: newproj\n")
	}
}

func TestProjectUpdate(t *testing.T) {
	p, _ := OpenProjectDB(testData)
	proj, _ := p.Find("proj1")

	p.Update(func() *Project {
		proj.Status = "Loaded"
		return proj
	})
	proj, _ = p.Find("proj1")
	if proj.Status != "Loaded" {
		t.Fatalf("proj1 status is %s, should have been 'Loaded'", proj.Status)
	}

	p2, _ := OpenProjectDB(testData)
	proj, _ = p2.Find("proj1")
	if proj.Status != "Loaded" {
		t.Fatalf("proj1 status is %s, should have been 'Loaded'", proj.Status)
	}
}
