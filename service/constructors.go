package service

// NewUsers creates a new Users instance connecting to a specific database backend.
func NewUsers(serviceDatabase ServiceDatabase) Users {
	switch serviceDatabase {
	case RethinkDB:
		return newRUsers()
	case SQL:
		panic("SQL ServiceDatabase not supported")
	default:
		panic("Unknown service type")
	}
}

// NewDirs creates a new Dirs instance connecting to a specific database backend.
func NewDirs(serviceDatabase ServiceDatabase) Dirs {
	switch serviceDatabase {
	case RethinkDB:
		return newRDirs()
	case SQL:
		panic("SQL ServiceDatabase not supported")
	default:
		panic("Unknown service type")
	}
}

// NewFiles creates a new Files instance connecting to a specific database backend.
func NewFiles(serviceDatabase ServiceDatabase) Files {
	switch serviceDatabase {
	case RethinkDB:
		return newRFiles()
	case SQL:
		panic("SQL ServiceDatabase not supported")
	default:
		panic("Unknown service type")
	}
}

// NewProjects creates a new Projects instance connecting to a specific database backend.
func NewProjects(serviceDatabase ServiceDatabase) Projects {
	switch serviceDatabase {
	case RethinkDB:
		return newRProjects()
	case SQL:
		panic("SQL ServiceDatabase not supported")
	default:
		panic("Unknown service type")
	}
}

// NewGroups creates a new Groups instance connecting to a specific database backend.
func NewGroups(serviceDatabase ServiceDatabase) Groups {
	switch serviceDatabase {
	case RethinkDB:
		return newRGroups()
	case SQL:
		panic("SQL ServiceDatabase not supported")
	default:
		panic("Unknown service type")
	}
}
