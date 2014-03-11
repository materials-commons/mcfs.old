package service

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
