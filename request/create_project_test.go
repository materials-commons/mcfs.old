package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/mc"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

func TestCreateProject(t *testing.T) {
	h := NewReqHandler(nil, session, "")
	h.user = "test@mc.org"

	createProjectRequest := protocol.CreateProjectReq{
		Name: "TestProject1__",
	}

	// Test create new project
	resp, err := h.createProject(&createProjectRequest)

	projectID := resp.ProjectID
	datadirID := resp.DataDirID

	if err != nil {
		t.Fatalf("Unable to create project")
	}

	// Make sure the created project is properly setup
	proj, err := model.GetProject(projectID, session)
	if err != nil {
		t.Errorf("Unable to retrieve project %s", projectID)
	}

	if proj.Name != "TestProject1__" {
		t.Errorf("Project Name not set")
	}

	if proj.DataDir == "" {
		t.Errorf("Project doesn't have a datadir associated with it")
	}

	// Make sure the join table is updated
	var p2d schema.Project2DataDir
	rql := r.Table("project2datadir").GetAllByIndex("project_id", projectID)
	err = model.GetRow(rql, session, &p2d)
	if err != nil {
		t.Errorf("Unable to find project in join table")
	}

	if p2d.DataDirID != datadirID {
		t.Errorf("Wrong datadir for project %#v expected %s", p2d, datadirID)
	}

	// Test create existing project
	resp, err = h.createProject(&createProjectRequest)
	if err != mc.ErrExists {
		t.Errorf("Creating an existing project should have returned err mc.ErrExists, returned %s instead", err)
	}

	// Delete before test so we can cleanup if there is a failure
	model.Delete("datadirs", datadirID, session)
	model.Delete("projects", projectID, session)
	model.Delete("project2datadir", p2d.ID, session)

	if err == nil {
		t.Fatalf("Created an existing project - shouldn't be able to")
	}

	if resp == nil {
		t.Fatalf("Creating an existing project should have returned its project id and datadir id")
	}

	if resp.ProjectID != projectID {
		t.Errorf("Creating an existing project returned the wrong project id")
	}

	if resp.DataDirID != datadirID {
		t.Errorf("Creating an existing project returned the wrong datadir id")
	}
	// Test create project with invalid name
	createProjectRequest.Name = "/InvalidName"
	resp, err = h.createProject(&createProjectRequest)
	if err == nil {
		t.Fatalf("Created project with Invalid name")
	}
}
