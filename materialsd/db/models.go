package db

import (
	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/mcfs/materialsd/db/model"
	"github.com/materials-commons/mcfs/materialsd/db/schema"
)

var (
	// projectsModel is the model for projects
	projectsModel *model.Model

	// projectEventsModel is the model for project events
	projectEventsModel *model.Model

	// projectFilesModel is the model for project files
	projectFilesModel *model.Model

	// Projects is the query model for projects
	Projects *model.Query

	// ProjectEvents is the query model for project events
	ProjectEvents *model.Query

	// ProjectFiles is the query model for project files
	ProjectFiles *model.Query
)

// Use sets the database connection for all the models.
func Use(db *sqlx.DB) {
	Projects = projectsModel.Q(db)
	ProjectEvents = projectEventsModel.Q(db)
	ProjectFiles = projectFilesModel.Q(db)
}

func init() {
	pQueries := model.ModelQueries{
		Insert: "insert into projects (name, path, mcid) values (:name, :path, :mcid)",
	}
	projectsModel = model.New(schema.Project{}, "projects", pQueries)

	peQueries := model.ModelQueries{
		Insert: `insert into project_events (path, event, event_time, project_id)
                 values (:path, :event, :event_time, :project_id)`,
	}
	projectEventsModel = model.New(schema.ProjectEvent{}, "project_events", peQueries)

	pfQueries := model.ModelQueries{
		Insert: `insert into project_files (path, size, checksum, mtime, atime, ctime, ftype, project_id, fidhigh, fidlow)
                 values (:path, :size, :checksum, :mtime, :atime, :ctime, :ftype, :project_id, :fidhigh, :fidlow)`,
	}
	projectFilesModel = model.New(schema.ProjectFile{}, "project_files", pfQueries)
}
