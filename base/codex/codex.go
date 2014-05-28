package codex

import (
	"bytes"
)

// PreparedBuffer represents a buffer stream with the first two
// bytes (type and version) removed.
type PreparedBuffer struct {
	Type    uint8
	Version uint8
	Bytes   []byte
}

// Encoder is the interface implemented by an object that can encode itself and
// tag its type and version.
type Encoder interface {
	Encode(otype uint8, version uint8, in interface{}) (*bytes.Buffer, error)
}

// Decoder is the interface implemented by an object that can be decoded after
// stripping out its tag type and version.
type Decoder interface {
	Decode(buf []byte, out interface{}) error
	Prepare(buf []byte) (*PreparedBuffer, error)
}

// EncoderDecoder implements both encoding and decoding.
type EncoderDecoder interface {
	Encoder
	Decoder
}
