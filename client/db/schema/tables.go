package schema

import (
	"database/sql"
	"fmt"
)

type schemaCommand struct {
	description string
	create      string
}

var schemas = []schemaCommand{
	{
		description: "Project Schema",
		create: `
                create table projects (
                   id   integer primary key,
                   name text,
                   path text,
                   mcid varchar(40)
                )`,
	},
	{
		description: "Project Changes Schema",
		create: `
                create table project_events (
                   id         integer primary key,
                   project_id integer,
                   path       text,
                   event      varchar(40),
                   event_time datetime,
                   foreign key (project_id) references projects(id)     
                )`,
	},
	{
		description: "Project Files Schema",
		create: `
                create table project_files (
                   id         integer primary key,
                   project_id integer,
                   path       text,
                   size       bigint,
                   checksum   varchar(32),
                   mtime      datetime,
                   atime      datetime,
                   ctime      datetime,
                   ftype      varchar(32),
                   fidhigh     int64,
                   fidlow      int64,
                   foreign key (project_id) references projects(id)
                )`,
	},
}

// Create creates the sql database by creating the tables and triggers.
func Create(db *sql.DB) error {
	for _, schema := range schemas {
		_, err := db.Exec(schema.create)
		if err != nil {
			return fmt.Errorf("failed on create for %s: %s", schema.description, err)
		}
	}

	return nil
}
