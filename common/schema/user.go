package schema

import (
	"time"
)

// ProjectPreferences is the users preferences for a particular project.
type ProjectPreferences struct {
}

// UserPreferences contains all the users preferences.
type UserPreferences struct {
	Templates      []string                      `gorethink:"templates"`
	DefaultProject string                        `gorethink:"default_project"`
	Projects       map[string]ProjectPreferences `gorethink:"projects"`
}

// User models a user in the system.
type User struct {
	ID          string          `gorethink:"id,omitempty"`
	Name        string          `gorethink:"name"`
	Email       string          `gorethink:"email"`
	Fullname    string          `gorethink:"fullname"`
	Password    string          `gorethink:"password"`
	APIKey      string          `gorethink:"apikey"`
	Birthtime   time.Time       `gorethink:"birthtime"`
	MTime       time.Time       `gorethink:"mtime"`
	Avatar      string          `gorethink:"avatar"`
	Description string          `gorethink:"description"`
	Affiliation string          `gorethink:"affiliation"`
	HomePage    string          `gorethink:"homepage"`
	Preferences UserPreferences `gorethink:"preferences"`
	Notes       []Note          `gorethink:"notes"`
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

// Clone clones a User object. It is currently a partial clone and doesn't
// copy over Notes or Preferences. This function is meant to be used
// where the notes and preferences are not needed.
func (u *User) Clone() *User {
	return &User{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Fullname:    u.Fullname,
		Password:    u.Password,
		APIKey:      u.APIKey,
		Birthtime:   u.Birthtime,
		MTime:       u.MTime,
		Avatar:      u.Avatar,
		Description: u.Description,
		Affiliation: u.Affiliation,
		HomePage:    u.HomePage,
	}
}

// IsValid validates
func (u *User) IsValid(id, apikey string) bool {
	return u.ID == id && u.APIKey == apikey
}
