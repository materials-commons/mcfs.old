package dai

import (
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/base/model"
	"github.com/materials-commons/mcfs/base/schema"
)

// rGroups implements the Groups interface for RethinkDB
type rGroups struct {
	session *r.Session
}

// newRGroups creates a new instance of rGroups
func newRGroups(session *r.Session) rGroups {
	return rGroups{
		session: session,
	}
}

// ByID looks up a group by its primary key.
func (g rGroups) ByID(id string) (*schema.Group, error) {
	var group schema.Group
	if err := model.Groups.Qs(g.session).ByID(id, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// Insert creates a new group.
func (g rGroups) Insert(group *schema.Group) (*schema.Group, error) {
	var newGroup schema.Group
	if err := model.Groups.Qs(g.session).Insert(group, &newGroup); err != nil {
		return nil, err
	}
	return &newGroup, nil
}

// Delete deletes a group. It updates all dependent objects.
func (g rGroups) Delete(id string) error {
	return model.Groups.Qs(g.session).Delete(id)
}

// HasAccess checks to see if the user making the request has access to the
// particular item. Access is determined as follows:
// 1. If the user and the owner of the item are the same return true (has access).
// 2. Get a list of all the users groups for the item's owner.
//    For each user in the user group see if the requesting user
//    is included. If so then return true (has access).
// 3. None of the above matched - return false (no access)
func (g rGroups) HasAccess(owner, user string) bool {
	return g.ownerGaveAccessTo(owner, user)
}

// ownerGaveAccessTo implements the algorithm described above for HasAccess
func (g rGroups) ownerGaveAccessTo(owner, user string) bool {
	// Check if user and file owner are the same, or the user is
	// in the admin group.
	if user == owner || g.isAdmin(user) {
		return true
	}

	// Get the owners groups
	rql := model.Groups.T().GetAllByIndex("owner", owner)
	var groups []schema.Group
	if err := model.Groups.Qs(g.session).Rows(rql, &groups); err != nil {
		// Some sort of error occurred, assume no access
		return false
	}

	// For each group go through its list of users and see if
	// they match the requesting user. If there is a match
	// then the owner has given access to the user.
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
func (g rGroups) isAdmin(user string) bool {
	group, err := g.ByID("admin")
	if err != nil {
		return false
	}

	for _, admin := range group.Users {
		if admin == user {
			return true
		}
	}

	return false
}
