package codex

import (
	"github.com/materials-commons/mcfs/base/protocol"
	"testing"
)

func TestEncode(t *testing.T) {
	m := NewMsgPak()

	lr := protocol.LoginReq{
		User:   "test@mc.org",
		APIKey: "test",
	}

	b, err := m.Encode(protocol.LoginRequest, 1, &lr)

	if err != nil {
		t.Fatalf("Unable to encode %#v: %s", lr, err)
	}

	bytesArray := b.Bytes()

	// Check that the message was encoded
	if uint8(bytesArray[0]) != LoginRequest {
		t.Fatalf("First byte doesn't encode LoginRequest(%d), instead it encodes %d", LoginRequest, bytesArray[0])
	}

	if uint8(bytesArray[1]) != 1 {
		t.Fatalf("Second bytes doesn't encode version (1), instead it encodes %d", bytesArray[1])
	}
}

func TestDecode(t *testing.T) {
	m := NewMsgPak()

	lr := protocol.LoginReq{
		User:   "test@mc.org",
		APIKey: "test",
	}

	b, err := m.Encode(protocol.LoginRequest, 1, &lr)
	if err != nil {
		t.Fatalf("Unable to encode %#v: %s", lr, err)
	}
	bytesArray := b.Bytes()

	pb, err := m.Prepare(bytesArray)
	if err != nil {
		t.Fatalf("Prepare bytes failed with %s", err)
	}

	if pb.Type != protocol.LoginRequest {
		t.Fatalf("Prepare contains wrong type expected %d, got %d", LoginRequest, pb.Type)
	}

	if pb.Version != 1 {
		t.Fatalf("PreparedBuffer contains wrong version expected %d, got %d", 1, pb.Version)
	}

	var lr2 protocol.LoginReq

	err = m.Decode(pb.Bytes, &lr2)
	if err != nil {
		t.Fatalf("Unable to decode bytes to a LoginReq: %s", err)
	}

	if lr2.User != "test@mc.org" {
		t.Fatalf("Decoded LoginRequest expected User 'test@mc.org', got '%s'", lr2.User)
	}

	if lr2.APIKey != "test" {
		t.Fatalf("Decoded LoginRequest expected APIKey 'test', got '%s'", lr2.APIKey)
	}
}
