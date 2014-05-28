package schema

import (
	"time"
)

// FType is the type of file
type FType int32

const (
	// FTypeFile File entry
	FTypeFile FType = iota // File

	// FTypeDirectory Directory entry
	FTypeDirectory

	// FTypeLink Soft link
	FTypeLink
)

// A Project is an instance of a users project.
type Project struct {
	ID   int    // Primary key
	Name string // Name of project
	Path string // Path to project
	MCID string // Materials Commons id for project
}

// A ProjectEvent is a file change event in the project.
type ProjectEvent struct {
	ID        int       // Primary key
	ProjectID int       `db:"project_id"` // Foreign key to project
	Path      string    // Path of file/directory this event pertains to
	Event     string    // Type of event
	EventTime time.Time `db:"event_time"` // Time event occurred
}

// A ProjectFile is a file or directory entry in the project. The type
// of entry is represented in the FType field. This currently supports
//
type ProjectFile struct {
	ID        int       // Primary key
	ProjectID int       `db:"project_id"` // Foreign key to project
	Path      string    // Full path to file/directory
	Size      int64     // Size of file (valid only for files)
	Checksum  string    // MD5 Hash of file (valid only for files)
	MTime     time.Time // Last known Modification time
	ATime     time.Time // Last access time
	CTime     time.Time // Creation time
	FType     string    // Type of entry
	FIDHigh   int64     // file.FID.IDHigh
	FIDLow    int64     // file.FID.IDLow
}
