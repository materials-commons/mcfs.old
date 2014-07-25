package tdb

import (
	"fmt"

	r "github.com/dancannon/gorethink"
)

// NewSession creates a new connection to the database. It panics
// if it cannot connect to the database.
func NewSession() *r.Session {
	s, err := r.Connect(r.ConnectOpts{
		Address:  "localhost:30815",
		Database: "materialscommons",
	})

	if err != nil {
		panic(fmt.Sprintf("Unable to connect to database: %s", err))
	}

	return s
}
