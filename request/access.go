package request

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/base/model"
	"github.com/materials-commons/base/schema"
)

// OwnerGaveAccessTo checks to see if the user making the request has access to the
// particular datafile. Access is determined as follows:
// 1. if the user and the owner of the file are the same return true (has access).
// 2. Get a list of all the users groups for the file owner.
//    For each user in the user group see if teh requesting user
//    is included. If so then return true (has access).
// 3. None of the above matched - return false (no access)
func OwnerGaveAccessTo(owner, user string, session *r.Session) bool {
	// Check if user and file owner are the same
	if user == owner {
		return true
	}

	if isAdmin(user, session) {
		return true
	}

	// Get the file owners usergroups
	rql := r.Table("usergroups").GetAllByIndex("owner", owner)
	groups, err := model.MatchingGroups(rql, session)
	if err != nil {
		return false
	}

	// For each usergroup go through its list of users
	// and see if they match the requesting user
	for _, group := range groups {
		users := group.Users
		for _, u := range users {
			if u == user {
				return true
			}
		}
	}

	return false
}

// isAdmin check if user is in admin table
func isAdmin(user string, session *r.Session) bool {
	var group schema.Group
	if err := model.GetItem("admin", "usergroups", session, &group); err != nil {
		return false
	}
	for _, admin := range group.Users {
		if admin == user {
			return true
		}
	}

	return false
}
