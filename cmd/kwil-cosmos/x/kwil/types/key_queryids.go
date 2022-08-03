package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// QueryidsKeyPrefix is the prefix to retrieve all Queryids
	QueryidsKeyPrefix = "Queryids/value/"
)

// QueryidsKey returns the store key to retrieve a Queryids from the index fields
func QueryidsKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
