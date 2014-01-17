package request

import (
	"fmt"
	"github.com/materials-commons/materials/util"
	"github.com/materials-commons/mcfs/protocol"
	"io"
	"testing"
)

var _ = fmt.Println

func TestReq(t *testing.T) {
	m := util.NewRequestResponseMarshaler()
	h := NewReqHandler(m, session, "")

	m.SetError(io.EOF)
	switch h.req().(type) {
	case protocol.CloseReq:
	default:
		t.Fatalf("Wrong type")
	}

	m.SetError(fmt.Errorf(""))
	switch h.req().(type) {
	case ErrorReq:
	default:
		t.Fatalf("Wrong type")
	}

	m.ClearError()
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
}
