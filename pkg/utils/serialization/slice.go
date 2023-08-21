package serialization

import (
	"encoding"
	"encoding/binary"
	"fmt"
)

// SerializeSlice serializes a slice of binary marshalable items.
// It prepends the length of each item to the serialized data.
func SerializeSlice[T encoding.BinaryMarshaler](slice []T) ([]byte, error) {
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

// DeserializeSlice deserializes a slice of binary marshalable items.
// It expects bytes, as well as a function to create a new item of the slice type.
func DeserializeSlice[T encoding.BinaryUnmarshaler](bts []byte, newFn func() T) ([]T, error) {
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
