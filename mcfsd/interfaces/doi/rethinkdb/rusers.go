package rethinkdb

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/interfaces/db/model"
	"github.com/materials-commons/mcfs/interfaces/db/schema"
)

// rUsers implements the Users interface for RethinkDB
type rUsers struct {
	session *r.Session
}

// NewRUsers creates a new instance of the rUsers for RethinkDB
func NewRUsers(session *r.Session) rUsers {
	return rUsers{
		session: session,
	}
}

// ByID looks up users by their primary key. In RethinkDB this is the id field.
func (u rUsers) ByID(id string) (*schema.User, error) {
	var user schema.User
	if err := model.Users.Qs(u.session).ByID(id, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ByAPIKey looks up users by their apikey. In RethinkDB this is the apikey field.
func (u rUsers) ByAPIKey(apikey string) (*schema.User, error) {
	var user schema.User
	rql := model.Users.T().GetAllByIndex("apikey", apikey)
	if err := model.Users.Qs(u.session).Row(rql, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// All returns all the users in the database.
func (u rUsers) All() ([]schema.User, error) {
	var users []schema.User
	if err := model.Users.Qs(u.session).Rows(model.Users.T(), &users); err != nil {
		return nil, err
	}
	return users, nil
}
