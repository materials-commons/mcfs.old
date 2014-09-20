package db

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/mcfs/materialsd/db/schema"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
	"time"
)

var _ = fmt.Println

var tdb *sqlx.DB

func init() {
	os.RemoveAll("/tmp/sqltest.db")
	var err error
	dbArgs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", "/tmp/sqltest.db")
	db, err := sql.Open("sqlite3", dbArgs)
	if err != nil {
		panic(fmt.Sprintf("models_test: Couldn't open test db: %s", err))
	}
	err = schema.Create(db)
	if err != nil {
		panic(fmt.Sprintf("Couldn't create test db: %s", err))
	}
	db.Close()
	tdb, err = sqlx.Open("sqlite3", dbArgs)
	if err != nil {
		panic("Couldn't reopen db under sqlx")
	}
	Use(tdb)
}

func TestProjects(t *testing.T) {
	proj := schema.Project{
		Name: "testproject",
		Path: "/tmp/testproject",
		MCID: "abc123",
	}

	err := Projects.Insert(proj)
	if err != nil {
		t.Fatalf("Insert Project %#v into projects failed %s", proj, err)
	}

	projects := []schema.Project{}
	err = Projects.Select(&projects, "select * from projects")
	if err != nil {
		t.Fatalf("Select of projects failed: %s", err)
	}

	if len(projects) != 1 {
		t.Fatalf("Expected to get back 1 project and instead got back %d", len(projects))
	}

	proj.ID = 1 // Set the id because first entry will have id 1
	if projects[0] != proj {
		t.Fatalf("Inserted proj different than retrieved version: i/r %#v/%#v", proj, projects[0])
	}

	// Test retrieve a single project
	var p schema.Project
	err = Projects.Get(&p, "select * from projects where path=$1", proj.Path)
	if err != nil {
		t.Errorf("Unable to retrieve a single project: %s", err)
	}

	if p != proj {
		t.Errorf("Inserted object different from retrieved object i/r %#v/%#v", proj, p)
	}

	// Test retrieve non existing
	err = Projects.Get(&p, "select * from projects where path=$1", "/does/not/exist")
	if err == nil {
		t.Errorf("Retrieved non existing project got: %#v", p)
	}
}

func TestProjectEvents(t *testing.T) {
	event := schema.ProjectEvent{
		Path:      "/tmp/testproject/abc.txt",
		Event:     "Delete",
		EventTime: time.Now(),
		ProjectID: 1,
	}

	err := ProjectEvents.Insert(event)

	if err != nil {
		t.Fatalf("Insert ProjectEvent %#v into project_events failed %s", event, err)
	}

	events := []schema.ProjectEvent{}
	err = ProjectEvents.Select(&events, "select * from project_events")
	if err != nil {
		t.Fatalf("Select of project_events failed: %s", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected to get back 1 event and instead got back %d", len(events))
	}

	event.ID = 1 // we know the first id in the database
	// nil out times since they won't be equal
	if !event.EventTime.Equal(events[0].EventTime) {
		t.Fatalf("Inserted event time not equal to retrieved i/r %#v/%#v", event, events[0])
	}

	// set times on both ends since we cannot indirectly compare the times
	// through structure comparison
	now := time.Now()
	event.EventTime = now
	events[0].EventTime = now
	if event != events[0] {
		t.Fatalf("Inserted event different than retrieved version: i/r %#v/%#v", event, events[0])
	}
}

func TestProjectFiles(t *testing.T) {
	defer cleanupMT()

	d := time.Date(2000, time.November, 12, 12, 0, 0, 0, time.UTC)
	f := schema.ProjectFile{
		Path:      "/tmp/testproject/abc.txt",
		ProjectID: 1,
		MTime:     d,
		CTime:     d,
		ATime:     d,
		Checksum:  "abc123",
		Size:      10,
		FIDHigh:   20,
		FIDLow:    30,
	}

	err := ProjectFiles.Insert(f)
	if err != nil {
		t.Fatalf("Insert ProjectFile %#v into project_files failed %s", f, err)
	}

	files := []schema.ProjectFile{}
	err = ProjectFiles.Select(&files, "Select * from project_files")
	if err != nil {
		t.Fatalf("Select of project_files failed: %s", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected to get back 1 file, and instead got back %d", len(files))
	}

	if !f.MTime.Equal(files[0].MTime) {
		t.Fatalf("MTime for inserted object is different than retrieved object")
	}

	// Structure comparison. Times will be different because of pointers so, just set both to now
	// since we already compared the times.
	now := time.Now()
	f.MTime = now
	f.CTime = now
	f.ATime = now
	files[0].MTime = now
	files[0].CTime = now
	files[0].ATime = now
	f.ID = 1 // We know the first id inserted is one
	if f != files[0] {
		t.Fatalf("Inserted object %#v, different from retrieved %#v", f, files[0])
	}
}

func cleanupMT() {
	tdb.Close()
	os.RemoveAll("/tmp/sqltest.db")
}
