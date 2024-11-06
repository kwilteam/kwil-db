package serialize

// This file provides a SerializeSlice function to serialize a slice of any type
// that is a standard encoding.BinaryMarshaler. The bytes from this function may
// be unmarshaled into a slice of the same type with DeserializeSlice if that
// type implements the encoding.BinaryUnmarshaler interface.
//
// Such types include those defined in types/transactions, but these helpers can
// be used for any qualified types in other packages. e.g. engine/types.User and
// kv/atomic.keyValue.

import (
	"encoding"
	"encoding/binary"
	"fmt"
)

// serializeSlice serializes a slice of binary marshalable items.
// It prepends the length of each item to the serialized data.
func serializeSlice[T encoding.BinaryMarshaler](slice []T) ([]byte, error) {
	var result []byte
	for _, item := range slice {
		data, err := item.MarshalBinary()
		if err != nil {
			return nil, err
		}

		length := uint64(len(data))
		result = append(result, append(make([]byte, 8), data...)...)
		binary.BigEndian.PutUint64(result[len(result)-len(data)-8:len(result)-len(data)], length)
	}

	return result, nil
}

// deserializeSlice deserializes a slice of binary marshalable items.
// It expects bytes, as well as a function to create a new item of the slice type.
func deserializeSlice[T encoding.BinaryUnmarshaler](bts []byte, newFn func() T) ([]T, error) {
	var result []T
	for len(bts) > 0 {
		if len(bts) < 8 {
			return nil, fmt.Errorf("insufficient bytes")
		}
		length := binary.BigEndian.Uint64(bts[:8])
		if uint64(len(bts[8:])) < length {
			return nil, fmt.Errorf("invalid length")
		}
		bts = bts[8:]

		value := newFn()
		err := value.UnmarshalBinary(bts[:length])
		if err != nil {
			return nil, err
		}

		result = append(result, value)
		bts = bts[length:]
	}

	return result, nil
}
