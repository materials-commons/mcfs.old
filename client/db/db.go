package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/materials-commons/gohandy/file"
	"github.com/materials-commons/mcfs/client/config"
	"github.com/materials-commons/mcfs/client/db/schema"
	_ "github.com/mattn/go-sqlite3" // Implicit import of driver
	"path/filepath"
)

// Open sets up the models and the database. If the database doesn't exist
// then Open will create the database and it's schema.
func Open() error {
	// The first thing we do is check if the database exists. We need to
	// know this prior to opening the database, since the database open
	// call will create the file.
	exists := Exists()

	// Open the database.
	dbargs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", Path())
	db, err := sqlx.Open("sqlite3", dbargs)
	if err != nil {
		return err
	}

	// If the database didn't exist, then we need to
	// create the schema.
	if !exists {
		if err := schema.Create(db.DB); err != nil {
			return err
		}
	}

	// Tell our models to use this database connection.
	Use(db)
	return nil
}

// Exists returns true if the database exists. This is determined by looking for
// a database file entry.
func Exists() bool {
	return file.Exists(Path())
}

// Path returns the full path to the database file.
func Path() string {
	return filepath.Join(config.Config.User.DotMaterialsPath(), "materials.db")
}
