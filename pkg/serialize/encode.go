package serialize

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kwilteam/kwil-db/pkg/utils/serialization"
)

type SerializedData = []byte

type encodingType uint16

const (
	// it is very important that the order of the encoding types is not changed
	encodingTypeInvalid encodingType = iota
	encodingTypeRLP
)

var currentEncodingType = encodingTypeRLP

// Encode encodes the given value into a serialized data format.
func Encode(val any) (SerializedData, error) {
	var btsVal []byte
	var err error
	switch currentEncodingType {
	case encodingTypeRLP:
		btsVal, err = encodeRLP(val)
	default:
		return nil, fmt.Errorf("invalid encoding type: %d", currentEncodingType)
	}
	if err != nil {
		return nil, err
	}

	return addSerializedTypePrefix(currentEncodingType, btsVal)
}

// Decode decodes the given serialized data into the given value.
func Decode[T any](bts SerializedData) (*T, error) {
	encType, val, err := removeSerializedTypePrefix(bts)
	if err != nil {
		return nil, err
	}

	switch encType {
	case encodingTypeRLP:
		return decodeRLP[T](val)
	default:
		return nil, fmt.Errorf("invalid encoding type: %d", val)
	}
}

func EncodeSlice[T any](kvs []T) ([]byte, error) {
	marshaller := make([]*serialBinaryMarshaller[T], len(kvs))
	for i, kv := range kvs {
		marshaller[i] = &serialBinaryMarshaller[T]{kv}
	}
	return serialization.SerializeSlice[*serialBinaryMarshaller[T]](marshaller)
}

func DecodeSlice[T any](bts []byte) ([]*T, error) {
	marshaller, err := serialization.DeserializeSlice[*serialBinaryMarshaller[T]](bts, func() *serialBinaryMarshaller[T] {
		return &serialBinaryMarshaller[T]{}
	})
	if err != nil {
		return nil, err
	}

	result := make([]*T, len(marshaller))
	for i, m := range marshaller {
		result[i] = &m.val
	}

	return result, nil
}

// serialBinaryMarshaller is a helper struct that implements the BinaryMarshaler and BinaryUnmarshaler interfaces
type serialBinaryMarshaller[T any] struct {
	val T
}

func (m *serialBinaryMarshaller[T]) MarshalBinary() ([]byte, error) {
	return Encode(m.val)
}

func (m *serialBinaryMarshaller[T]) UnmarshalBinary(bts []byte) error {
	val, err := Decode[T](bts)
	if err != nil {
		return err
	}

	m.val = *val
	return nil
}

/*

// EncodeSlice serializes a slice into a byte slice
func EncodeSlice[T any](kvs []T) ([]byte, error) {
	var result []byte
	for _, kv := range kvs {
		data, err := Encode(kv)
		if err != nil {
			return nil, err
		}
		length := uint64(len(data))
		result = append(result, append(make([]byte, 8), data...)...)
		binary.BigEndian.PutUint64(result[len(result)-len(data)-8:len(result)-len(data)], length)
	}
	return result, nil
}

// DecodeSlice deserializes a byte slice into a slice
// It is important this is only given the results of EncodeSlice
func DecodeSlice[T any](bts []byte) ([]*T, error) {
	var result []*T
	for len(bts) > 0 {
		if len(bts) < 8 {
			return nil, fmt.Errorf("insufficient bytes")
		}
		length := binary.BigEndian.Uint64(bts[:8])
		if uint64(len(bts[8:])) < length {
			return nil, fmt.Errorf("invalid length")
		}
		bts = bts[8:]
		value, err := Decode[T](bts[:length])
		if err != nil {
			return nil, err
		}

		result = append(result, value)
		bts = bts[length:]
	}
	return result, nil
}
*/

func encodeRLP(val any) ([]byte, error) {
	return rlp.EncodeToBytes(val)
}

func decodeRLP[T any](bts []byte) (*T, error) {
	var val T
	err := rlp.DecodeBytes(bts, &val)
	if err != nil {
		return nil, err
	}

	return &val, nil
}

func DecodeInto(bts []byte, v any) error {
	encType, val, err := removeSerializedTypePrefix(bts)
	if err != nil {
		return err
	}

	switch encType {
	case encodingTypeRLP:
		return rlp.DecodeBytes(val, v)
	default:
		return fmt.Errorf("invalid encoding type: %d", val)
	}
}
