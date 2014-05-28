package schema

// Project2DataDir is a join table that maps projects to their datadirs.
type Project2DataDir struct {
	ID        string `gorethink:"id,omitempty" db:"-"`
	ProjectID string `gorethink:"project_id" db:"project_id"`
	DataDirID string `gorethink:"datadir_id" db:"datadir_id"`
}
