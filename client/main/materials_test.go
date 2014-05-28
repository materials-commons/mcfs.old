package main

import (
	"bitbucket.org/kardianos/osext"
	"github.com/materials-commons/mcfs/client"
	"github.com/materials-commons/mcfs/client/config"
	"github.com/materials-commons/mcfs/client/user"
	"path/filepath"
	"testing"
)

func TestConvertProjects(t *testing.T) {
	u, _ := user.NewUserFrom("../test_data/conversion")
	config.ConfigInitialize(u)
	convertProjects()

	// Make sure the conversion went correctly
	projectDB, err := materials.OpenProjectDB("../test_data/conversion/.materials/projectdb")

	if err != nil {
		t.Fatalf("Unable to open projectdb %s", err.Error())
	}

	for _, project := range projectDB.Projects() {
		switch {
		case project.Name == "proj1a":
			verify(project, "/tmp/proj1a", "Unloaded", t)
		case project.Name == "proj 2a":
			verify(project, "/tmp/proj 2a", "Loaded", t)
		default:
			t.Fatalf("Unexpected project %#v\n", project)
		}
	}
}

func verify(project *materials.Project, path, status string, t *testing.T) {
	if project.Path != path {
		t.Fatalf("Paths don't match, expected %s, got %s\n", project.Path, path)
	}

	if project.Status != status {
		t.Fatalf("Status don't match, expected %s, got %s\n", project.Status, status)
	}
}

func TestAddProject(t *testing.T) {
	u, _ := user.NewUserFrom("../test_data")
	config.ConfigInitialize(u)

	// Test add new project
	folderPath, _ := osext.ExecutableFolder()

	projectName := filepath.Base(folderPath)
	err := addProject(projectName, folderPath)
	if err != nil {
		t.Errorf("Unable to create a valid project %s:%s", projectName, folderPath)
	}

	// Test add existing project

	// Test add with blank project name

	// Test add with blank project path

	// Test add with path that doesn't exist

	// Test add relative path

	// Test add where last element of project path != project name
}
