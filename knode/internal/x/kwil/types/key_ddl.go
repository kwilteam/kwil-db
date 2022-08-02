package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// DdlKeyPrefix is the prefix to retrieve all Ddl
	DdlKeyPrefix = "Ddl/value/"
)

// DdlKey returns the store key to retrieve a Ddl from the index fields
func DdlKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
