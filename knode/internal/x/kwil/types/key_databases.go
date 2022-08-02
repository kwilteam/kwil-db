package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// DatabasesKeyPrefix is the prefix to retrieve all Databases
	DatabasesKeyPrefix = "Databases/value/"
)

// DatabasesKey returns the store key to retrieve a Databases from the index fields
func DatabasesKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
