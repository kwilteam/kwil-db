package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// DdlindexKeyPrefix is the prefix to retrieve all Ddlindex
	DdlindexKeyPrefix = "Ddlindex/value/"
)

// DdlindexKey returns the store key to retrieve a Ddlindex from the index fields
func DdlindexKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
