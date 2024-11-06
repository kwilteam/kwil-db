package types

import "kwil/node/types/serialize"

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
