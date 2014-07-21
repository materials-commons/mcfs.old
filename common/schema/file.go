package schema

import (
	"time"
)

// File models a user file. A datafile is an abstract representation of a real file
// plus the attributes that we need in our model for access, and other metadata.
type File struct {
	ID          string    `gorethink:"id,omitempty"` // Primary key.
	Current     bool      `gorethink:"current"`      // Is this the most current version.
	Name        string    `gorethink:"name"`         // Name of file.
	Birthtime   time.Time `gorethink:"birthtime"`    // Creation time.
	MTime       time.Time `gorethink:"mtime"`        // Modification time.
	ATime       time.Time `gorethink:"atime"`        // Last access time.
	Description string    `gorethink:"description"`
	Notes       []string  `gorethink:"notes"`
	Owner       string    `gorethink:"owner"`     // Who owns the file.
	Checksum    string    `gorethink:"checksum"`  // MD5 Hash.
	Size        int64     `gorethink:"size"`      // Size of file.
	Uploaded    int64     `gorethink:"uploaded"`  // Number of bytes uploaded. When Size != Uploaded file is only partially uploaded.
	MediaType   string    `gorethink:"mediatype"` // mime type.
	Parent      string    `gorethink:"parent"`    // If there are multiple ids then parent is the id of the previous version.
	UsesID      string    `gorethink:"usesid"`    // If file is a duplicate, then usesid points to the real file. This allows multiple files to share a single physical file.
	DataDirs    []string  `gorethink:"datadirs"`  // List of the directories the file can be found in.
}

// NewFile creates a new File instance.
func NewFile(name, owner string) File {
	now := time.Now()
	return File{
		Name:        name,
		Owner:       owner,
		Description: "",
		Birthtime:   now,
		MTime:       now,
		ATime:       now,
		Current:     true,
	}
}

// FileID returns the id to use for the file. Because files can be duplicates, all
// duplicates are stored under a single ID. UsesID is set to the ID that an entry
// points to when it is a duplicate.
func (f *File) FileID() string {
	if f.UsesID != "" {
		return f.UsesID
	}

	return f.ID
}

// private type to hang methods off of
type fs struct{}

// Files gives access to help routines that work on lists of files.
var Files fs

// Find will return a matching File in a list of files when the match func returns true.
func (f fs) Find(files []File, match func(f File) bool) *File {
	for _, file := range files {
		if match(file) {
			return &file
		}
	}

	return nil
}
