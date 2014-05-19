package request

import (
	"encoding/gob"
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/materials-commons/mcfs/protocol"
	"net"
	"os"
	"testing"
)

var session *r.Session

func init() {
	session, _ = r.Connect(map[string]interface{}{
		"address":  "localhost:30815",
		"database": "materialscommons",
	})
}

type client struct {
	*gob.Encoder
	*gob.Decoder
}

func newClient() *client {
	conn, err := net.Dial("tcp", "localhost:35862")
	if err != nil {
		fmt.Printf("Couldn't connect %s\n", err.Error())
		os.Exit(1)
	}
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)
	return &client{
		Encoder: encoder,
		Decoder: decoder,
	}
}

var gtarceaLoginReq = protocol.LoginReq{
	User:   "test@mc.org",
	APIKey: "test",
}

func loginTestUser() *client {
	client := newClient()
	request := protocol.Request{&gtarceaLoginReq}
	client.Encode(&request)
	resp := protocol.Response{}
	client.Decode(&resp)
	return client
}

func TestLoginLogout(t *testing.T) {
	h := NewReqHandler(nil, "")
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
	_, err = h.logout(&logoutRequest)
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
