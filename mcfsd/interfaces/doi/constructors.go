package doi

import (
	"fmt"

	"github.com/materials-commons/mcfs/interfaces/db"
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
			File:    newRFiles(session),
			Dir:     newRDirs(session),
			Project: newRProjects(session),
			Group:   newRGroups(session),
			User:    newRUsers(session),
		}
	case SQL:
		panic("SQL ServiceDatabase not supported")
	default:
		panic("Unknown service type")
	}
}
