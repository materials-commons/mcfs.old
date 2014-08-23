package mcfs

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/mcfs/client/db"
	"github.com/materials-commons/mcfs/client/db/schema"
	"github.com/materials-commons/mcfs/mcerr"
	"github.com/materials-commons/mcfs/model"
	"github.com/materials-commons/mcfs/server/inuse"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

var tdb *sqlx.DB

func init() {
	os.RemoveAll("/tmp/sqltest.db")
	var err error
	dbArgs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", "/tmp/sqltest.db")
	cdb, err := sql.Open("sqlite3", dbArgs)
	if err != nil {
		panic(fmt.Sprintf("project_test: couldn't open db: %s", err))
	}
	schema.Create(cdb)
	cdb.Close()
	tdb, err = sqlx.Open("sqlite3", dbArgs)
	if err != nil {
		panic("Couldn't reopen db under sqlx")
	}
	db.Use(tdb)
}

func TestProjectExistence(t *testing.T) {
	p := schema.Project{
		Name: "TestProject",
		Path: "/tmp/TestProject",
	}
	db.Projects.Insert(p)
	// Test Existing Project
	proj, err := projectByPath("/tmp/TestProject")
	if err != nil {
		t.Errorf("Failed to retrieve existing project")
	}
	p.ID = proj.ID
	if *proj != p {
		t.Errorf("Retrieve project differs from inserted version i/r %#v/%#v", p, proj)
	}

	// Test Non Existing Project
	proj, err = projectByPath("/tmp/TestProject-does-not-exist")
	if err == nil {
		t.Errorf("Successfully retrieve a non existing project")
	}

	// Test Project with Same name but different path (it should be found)
	proj, err = projectByPath("/does/not/exist/TestProject")
	if err != nil {
		t.Errorf("Failed to retrieve existing project")
	}
	if *proj != p {
		t.Errorf("Retrieve project differs from inserted version i/r %#v/%#v", p, proj)
	}
}

func TestUploadNewProject(t *testing.T) {
	// Test large upload
	if true {
		return
	}
	err := c.UploadNewProject("/home/gtarcea/ST1")
	if err != nil {
		t.Errorf("Failed to upload %s", err)
	}
}

func TestCreateProject(t *testing.T) {
	// Test Create New Project
	proj, err := c.CreateProject("NewProject")
	if err != nil {
		t.Errorf("Failed to create a new project")
	}

	projectID := proj.ProjectID
	dataDirID := proj.DataDirID

	// Test Create Existing Project

	// First unlock
	inuse.Unmark(projectID)
	proj2, err := c.CreateProject("NewProject")

	// Delete before testing
	model.Delete("projects", projectID, session)
	model.Delete("datadirs", dataDirID, session)

	if err != mcerr.ErrExists {
		t.Errorf("Creating an existing project should have returned mcerr.ErrExists: %s", err)
	}

	if proj2 == nil {
		t.Fatalf("Create existing project should have returned project")
	}

	if proj2.ProjectID != proj.ProjectID {
		t.Errorf("Create existing project returned wrong project id")
	}

	if proj2.DataDirID != proj.DataDirID {
		t.Errorf("Create existing project returned wrong datadir id")
	}
}
