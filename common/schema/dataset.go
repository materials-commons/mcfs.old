package schema

import "time"

type DataSet struct {
	ID        string    `gorethink:"id"`
	Birthtime time.Time `gorethink:"birthtime"`
	MTime     time.Time `gorethink:"mtime"`
	Owner     string    `gorethink:"owner"`
	Inbox     string    `gorethink:"inbox"`
	Files     []string  `gorethink:"files"`
	Tags      []string  `gorethink:"tags"`
}
