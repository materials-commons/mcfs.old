package schema

import (
	"time"
)

// User models a user in the system.
type User struct {
	ID          string    `gorethink:"id,omitempty"`
	Name        string    `gorethink:"name"`
	Email       string    `gorethink:"email"`
	Fullname    string    `gorethink:"fullname"`
	Password    string    `gorethink:"password"`
	APIKey      string    `gorethink:"apikey"`
	Birthtime   time.Time `gorethink:"birthtime"`
	MTime       time.Time `gorethink:"mtime"`
	Avatar      string    `gorethink:"avatar"`
	Description string    `gorethink:"description"`
	Affiliation string    `gorethink:"affiliation"`
	HomePage    string    `gorethink:"homepage"`
	Notes       []string  `gorethink:"notes"`
}

// NewUser creates a new User instance.
func NewUser(name, email, password, apikey string) User {
	now := time.Now()
	return User{
		ID:        email,
		Name:      name,
		Email:     email,
		Password:  password,
		APIKey:    apikey,
		Birthtime: now,
		MTime:     now,
	}
}
