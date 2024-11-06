package types

import (
	"fmt"
	"kwil/types/serialize"
)

// PayloadType is the type of payload
type PayloadType string

func (p PayloadType) String() string {
	return string(p)
}

// Payload is the interface that all payloads must implement
// Implementations should use Kwil's serialization package to encode and decode themselves
type Payload interface {
	MarshalBinary() (serialize.SerializedData, error)
	UnmarshalBinary(serialize.SerializedData) error
	Type() PayloadType
}

const (
	PayloadTypeKV PayloadType = "kv"
)

// payloadConcreteTypes associates a payload type with the concrete type of
// Payload. Use with UnmarshalPayload or reflect to instantiate.
var payloadConcreteTypes = map[PayloadType]Payload{
	PayloadTypeKV: &KVPayload{},
}

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

func (p *KVPayload) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(p)
}

func (p *KVPayload) UnmarshalBinary(data serialize.SerializedData) error {
	return serialize.Decode(data, p)
}

func (p *KVPayload) Type() PayloadType {
	return PayloadTypeKV
}
