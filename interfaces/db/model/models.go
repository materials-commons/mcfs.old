package model

import (
	"github.com/materials-commons/mcfs/common/schema"
	dbschema "github.com/materials-commons/mcfs/interfaces/db/schema"
)

// Groups is a default model for the usergroups table.
var Groups = &Model{
	schema: schema.Group{},
	table:  "usergroups",
}

// Users is a default model for the users table.
var Users = &Model{
	schema: schema.User{},
	table:  "users",
}

// Dirs is a default model for the datadirs table.
var Dirs = &Model{
	schema: schema.Directory{},
	table:  "datadirs",
}

// DirsDenorm is a default model for the denormalized datadirs_denorm table
var DirsDenorm = &Model{
	schema: dbschema.DataDirDenorm{},
	table:  "datadirs_denorm",
}

// Files is a default model for the datafiles table
var Files = &Model{
	schema: schema.File{},
	table:  "datafiles",
}

// Projects is a default model for the projects table
var Projects = &Model{
	schema: schema.Project{},
	table:  "projects",
}
