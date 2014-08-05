package schema

import "time"

type Access struct {
	ID          string    `gorethink:"id,omitempty"`
	ProjectID   string    `gorethink:"project_id"`
	ProjectName string    `gorethink:"project_name"`
	UserID      string    `gorethink:"user_id"`
	Birthtime   time.Time `gorethink:"birthtime"`
	MTime       time.Time `gorethink:"mtime"`
	Dataset     string    `gorethink:"dataset"`
	Permissions string    `gorethink:"permissions"`
}
