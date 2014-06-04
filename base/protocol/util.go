package protocol

import (
	"bytes"
	"github.com/materials-commons/mcfs/base/codex"
	"github.com/materials-commons/mcfs/base/mcerr"
	"github.com/ugorji/go/codec"
)

/*
The following encodes and decodes buffers of bytes using MessagePack. It uses the approach
found in github.com/hashicorp/serf for identifying the type of message. The buffer has
a message type prepended to it. In our implementation we also prepend a version so that
multiple protocol versions can be supported.
*/

// EncodeCurrentVersion encodes a message using MsgPack. It prepends the message type and
// the CurrentVersion to the returned buffer.
func EncodeCurrentVersion(msgType uint8, in interface{}) (*bytes.Buffer, error) {
	return Encode(msgType, CurrentVersion, in)
}

// Encode encodes a message using MessagePack. It prepends the message type and the passed in
// version to the returned buffer.
func Encode(msgType uint8, version uint8, in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(msgType)
	buf.WriteByte(version)
	handle := codec.MsgpackHandle{}
	encoder := codec.NewEncoder(buf, &handle)
	err := encoder.Encode(in)
	return buf, err
}

// Decode decodes a buffer using MessagePack. The buffer passed in needs to have removed the
// message type and version from the buf.
func Decode(buf []byte, out interface{}) error {
	reader := bytes.NewReader(buf)
	handle := codec.MsgpackHandle{}
	decoder := codec.NewDecoder(reader, &handle)
	return decoder.Decode(out)
}

// Prepare retrieves the message type, version, and a buffer that is ready to be
// sent to Decode.
func Prepare(buf []byte) (pb *codex.PreparedBuffer, err error) {
	if len(buf) < 3 {
		return nil, mcerr.ErrInvalid
	}

	pb = &codex.PreparedBuffer{
		Type:    uint8(buf[0]),
		Version: uint8(buf[1]),
		Bytes:   buf[2:],
	}
	return pb, nil
}
