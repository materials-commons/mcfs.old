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
