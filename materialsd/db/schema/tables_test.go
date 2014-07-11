package schema

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"testing"
)

var _ = fmt.Println

func TestCreate(t *testing.T) {
	dbArgs := fmt.Sprintf("file:%s?cached=shared&mode=rwc", "/tmp/sqltestc.db")
	db, err := sql.Open("sqlite3", dbArgs)
	if err != nil {
		t.Fatalf("Unable to open database %s", err)
	}
	err = Create(db)
	db.Close()
	defer cleanup("/tmp/sqltestc.db")

	if err != nil {
		t.Errorf("Unable to create database: %s", err)
	}

}

func cleanup(path string) {
	os.RemoveAll(path)
}
