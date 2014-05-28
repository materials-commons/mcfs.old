package db

import (
	r "github.com/dancannon/gorethink"
)

var dbAddress = ""
var dbName = ""

// SetAddress sets the address to connect to the RethinkDB database.
func SetAddress(address string) {
	dbAddress = address
}

// SetDatabase sets the default database to use.
func SetDatabase(db string) {
	dbName = db
}

// RSession creates a new RethinkDB session.
func RSession() (*r.Session, error) {
	return r.Connect(
		r.ConnectOpts{
			Address:   dbAddress,
			Database:  dbName,
			MaxIdle:   10,
			MaxActive: 20,
		})
}
