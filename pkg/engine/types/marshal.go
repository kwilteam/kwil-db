package types

import (
	"fmt"
	"kwil/pkg/utils/serialize"

	"github.com/cstockton/go-conv"
)

func marshal(v any, d DataType) ([]byte, error) {
	switch d {
	case NULL:
		return prepend(NULL, nil), nil
	case TEXT:
		str, err := conv.String(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to string: %v", err)
		}

		return prepend(TEXT, serialize.StringToBytes(str)), nil
	case INT:
		i, err := conv.Int64(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to int64: %v", err)
		}

		return prepend(INT, serialize.Int64ToBytes(int64(i))), nil
	}

	return nil, fmt.Errorf("unknown type: %d", d)
}

func prepend(b DataType, bts []byte) []byte {
	return append([]byte{byte(b)}, bts...)
}

// tryUnmarshal will try to unmarshal a byte slice to the specified type
func tryUnmarshal(bts []byte, d DataType) (any, error) {
	switch d {
	case NULL:
		return nil, nil
	case TEXT:
		return serialize.BytesToString(bts), nil
	case INT:
		return serialize.BytesToInt64(bts), nil
	}

	return nil, fmt.Errorf("unknown type: %d", d)
}
