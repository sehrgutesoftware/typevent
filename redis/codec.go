package redis

import (
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

// defaultCodec is the default codec used to marshal and unmarshal events.
var defaultCodec = &MsgpackCodec{}

// the following are compile time assertions to ensure the codecs implement the interface.
var (
	_ Codec = &MsgpackCodec{}
	_ Codec = &JSONCodec{}
)

// Codec is the Codec used to marshal and unmarshal events.
type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// MsgpackCodec is a codec that uses msgpack as the encoding.
type MsgpackCodec struct{}

// MarshalBinary turns the Event into a binary representation using msgpack.
func (c *MsgpackCodec) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

// UnmarshalBinary unmarshals a binary representation of the Event using msgpack.
func (c *MsgpackCodec) Unmarshal(data []byte, v any) error {
	return msgpack.Unmarshal(data, v)
}

// JSONCodec is a codec that uses JSON as the encoding.
type JSONCodec struct{}

// Marshal turns the Event into a binary representation using JSON.
func (c *JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal unmarshals a binary representation of the Event using JSON.
func (c *JSONCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
