package types

import (
	"bytes"
	"encoding"
	"fmt"
)

// PayloadType is the type of payload
type PayloadType string

func (p PayloadType) String() string {
	return string(p)
}

// Payload is the interface that all payloads must implement
// Implementations should use Kwil's serialization package to encode and decode themselves
type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	Type() PayloadType
}

const (
	PayloadTypeKV PayloadType = "kv"
)

// payloadTypes includes native types and types registered from extensions.
var payloadTypes = map[PayloadType]bool{
	PayloadTypeKV: true,
}

// Valid says if the payload type is known. This does not mean that the node
// will execute the transaction, e.g. not yet activated, or removed.
func (p PayloadType) Valid() bool {
	// native types first for speed
	switch p {
	case PayloadTypeKV:
		return true
	default:
		return payloadTypes[p]
	}
}

// RegisterPayload registers a new payload type. This should be done on
// application initialization. A known payload type does not require a
// corresponding route handler to be registered with extensions/consensus so
// that they become available for consensus according to chain config.
func RegisterPayload(pType PayloadType) {
	if _, have := payloadTypes[pType]; have {
		panic(fmt.Sprintf("already have payload type %v", pType))
	}
	payloadTypes[pType] = true
}

// KVPair payload for testing purposes
type KVPayload struct {
	Key   []byte
	Value []byte
}

var _ Payload = &KVPayload{}

var _ encoding.BinaryMarshaler = (*KVPayload)(nil)
var _ encoding.BinaryMarshaler = KVPayload{}

func (p KVPayload) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(lenSize + len(p.Key) + lenSize + len(p.Value))
	err := writeBytes(&buf, p.Key)
	if err != nil {
		return nil, err
	}
	err = writeBytes(&buf, p.Value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ encoding.BinaryUnmarshaler = (*KVPayload)(nil)

func (p *KVPayload) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	kb, err := readBytes(r)
	if err != nil {
		return err
	}
	vb, err := readBytes(r)
	if err != nil {
		return err
	}
	// if r.Len() > 0 {
	// 	return fmt.Errorf("extra bytes after kv payload")
	// }

	p.Key = kb
	p.Value = vb

	return nil
}

func (p *KVPayload) Type() PayloadType {
	return PayloadTypeKV
}
