package request

import (
	"fmt"
	"github.com/materials-commons/materials/util"
	"github.com/materials-commons/contrib/mc"
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

func TestSsfAndSs(t *testing.T) {
	s := ssf(mc.ErrorCodeInvalid, "Error %s", "a")
	if s.err.Error() != "Error a" {
		t.Errorf("error string wrong: %s", s.err)
	}

	s = ss(mc.ErrorCodeInvalid, mc.ErrInvalid)
	if s.err != mc.ErrInvalid {
		t.Errorf("Not equal to mc.ErrInvalid: %s", s.err)
	}
}
