package request

import (
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/codex"
	"github.com/materials-commons/mcfs/protocol"
	"testing"
)

var _ = fmt.Println

var session *r.Session

func init() {
	session, _ = r.Connect(
		r.ConnectOpts{
			Address:  "localhost:30815",
			Database: "materialscommons",
		})
}

func TestLoginLogout(t *testing.T) {
	h := NewReqHandler(nil, codex.NewMsgPak(), "")
	h.user = "test@mc.org"

	// Test valid login
	loginRequest := protocol.LoginReq{
		User:   "test@mc.org",
		APIKey: "test",
	}

	_, err := h.login(&loginRequest)
	if err != nil {
		t.Fatalf("Failed to login with valid user id %s", err)
	}

	// Test logout
	logoutRequest := protocol.LogoutReq{}
	err = h.logout(&logoutRequest)
	if err != nil {
		t.Fatalf("logout failed %s", err)
	}

	// Test Bad Apikey with a known user
	loginRequest.APIKey = "abc12356"
	_, err = h.login(&loginRequest)
	if err == nil {
		t.Fatalf("Successful login with bad apikey")
	}

	// Test good Apikey with wrong user
	loginRequest.APIKey = "test2-wrong"
	loginRequest.User = "test2@mc.org"
	_, err = h.login(&loginRequest)
	if err == nil {
		t.Fatalf("Login successful with good api but wrong user")
	}

	// Test good Apikey with an non existing user
	loginRequest.User = "i@donotexist.com"
	_, err = h.login(&loginRequest)
	if err == nil {
		t.Fatalf("Login successful with good api but a non existing user")
	}

}
