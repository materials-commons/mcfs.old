package request

import (
	"fmt"
	"github.com/materials-commons/mcfs/base/mc"
	"github.com/materials-commons/mcfs/client/util"
	"github.com/materials-commons/mcfs/server/protocol"
	"testing"
)

var _ = fmt.Println

func TestRequestMarshaler(t *testing.T) {
	m := util.NewRequestResponseMarshaler()
	request := protocol.Request{1}
	m.Marshal(&request)
	var d protocol.Request
	if err := m.Unmarshal(&d); err != nil {
		t.Fatalf("Unmarshal failed with error %s", err)
	}

	if d.Req != 1 {
		t.Fatalf("Inner item not being properly saved")
	}
}

func TestChannelMarshaler(t *testing.T) {
	m := util.NewChannelMarshaler()
	go responder(m)
	loginReq := protocol.LoginReq{
		User:   "gtarcea@umich.edu",
		APIKey: "abc123",
	}
	req := protocol.Request{
		Req: loginReq,
	}

	if true {
		return
	}
	m.Marshal(&req)
	var resp protocol.Response
	m.Unmarshal(&resp)
	fmt.Printf("resp = %#v\n", resp)
	l := resp.Resp.(*protocol.LogoutResp)
	fmt.Printf("l = %#v\n, *l = %#v\n", l, *l)
}

func responder(m *util.ChannelMarshaler) {
	var request protocol.Request
	m.Unmarshal(&request)
	fmt.Printf("request = %#v\n", request)
	logoutResp := protocol.LogoutResp{}
	resp := protocol.Response{
		Status: mc.ErrorCodeSuccess,
		Resp:   &logoutResp,
	}
	m.Marshal(&resp)
}
