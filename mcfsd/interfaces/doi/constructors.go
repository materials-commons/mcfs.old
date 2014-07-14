package doi

import (
	"fmt"

	"github.com/materials-commons/mcfs/interfaces/db"
	"github.com/materials-commons/mcfs/mcfsd/interfaces/doi/rethinkdb"
)

type Service struct {
	File    Files
	Dir     Dirs
	Project Projects
	Group   Groups
	User    Users
}

func New(serviceDatabase ServiceDatabase) *Service {
	switch serviceDatabase {
	case RethinkDB:
		session, err := db.RSession()
		if err != nil {
			panic(fmt.Sprintf("Unable to connect to database: %s", err))
		}
		return &Service{
			File:    rethinkdb.NewRFiles(session),
			Dir:     rethinkdb.NewRDirs(session),
			Project: rethinkdb.NewRProjects(session),
			Group:   rethinkdb.NewRGroups(session),
			User:    rethinkdb.NewRUsers(session),
		}
	case SQL:
		panic("SQL ServiceDatabase not supported")
	default:
		panic("Unknown service type")
	}
}
