package request

import (
	"fmt"
	"testing"
)

var _ = fmt.Println

func TestReq(t *testing.T) {
	/*
		h := NewReqHandler(nil, codex.NewMsgPak(), "")

		switch h.req().(type) {
		case protocol.CloseReq:
		default:
			t.Fatalf("Wrong type")
		}

		switch h.req().(type) {
		case errorReq:
		default:
			t.Fatalf("Wrong type")
		}

		loginReq := protocol.LoginReq{}
		request := protocol.Request{loginReq}
		if err := m.Marshal(&request); err != nil {
			t.Fatalf("Marshal failed")
		}
		val := h.req()
		switch val.(type) {
		case protocol.LoginReq:
		default:
			t.Fatalf("req returned wrong type %T", val)
		}
	*/
}
