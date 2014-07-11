package access

import (
	"fmt"
	"github.com/materials-commons/mcfs/db"
	"github.com/materials-commons/mcfs/model"
	"github.com/materials-commons/mcfs/schema"
	"github.com/materials-commons/mcfs/server"
	"github.com/materials-commons/mcfs/server/service"
	"testing"
)

var _ = fmt.Println

func init() {
	mcfs.InitRethinkDB()
}

var _apikeys = newAPIKeys(service.New(service.RethinkDB).User)

func TestLoad(t *testing.T) {
	db.SetAddress("localhost:30815")
	db.SetDatabase("materialscommons")
	err := _apikeys.load()
	if err != nil {
		t.Fatalf("Loading apikeys failed %s", err)
	}

	if len(_apikeys.keys) == 0 {
		t.Fatalf("No apikeys")
	}
}

func TestReloadAndLookup(t *testing.T) {
	// Add a user, reload and make sure that the user loaded.
	user := schema.NewUser("tuser", "tuser@test.org", "abc123", "apikey123")
	model.Users.Q().Insert(user, nil)
	defer func() {
		model.Users.Q().Delete("tuser@test.org")
	}()
	_apikeys.reload()
	_, found := _apikeys.lookup("apikey123")
	if !found {
		t.Fatalf("apikeys reload failed, couldn't find user with key apikey123")
	}
	// Modify a user, reload and make sure that the user was modified
	user.APIKey = "abc123"
	model.Users.Q().Update(user.ID, user)
	_apikeys.reload()
	_, found = _apikeys.lookup("apikey123")
	if found {
		t.Fatalf("apikeys found key just replaced")
	}

	_, found = _apikeys.lookup("abc123")
	if !found {
		t.Fatalf("new apikey abc123 not found")
	}
}
