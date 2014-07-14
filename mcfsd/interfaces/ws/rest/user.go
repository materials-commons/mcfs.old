package rest

// User contains the user authentication information sent with the request.
type User struct {
	Name  string
	Token string
}
