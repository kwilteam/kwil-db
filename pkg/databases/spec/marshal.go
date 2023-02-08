package spec

import (
	"fmt"
	"github.com/cstockton/go-conv"
	"kwil/pkg/utils/serialize"
)

func marshal(v any, d DataType) ([]byte, error) {
	switch d {
	case NULL:
		return prepend(NULL, nil), nil
	case STRING:
		str, err := conv.String(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to string: %v", err)
		}

		return prepend(STRING, serialize.StringToBytes(str)), nil
	case INT32:
		i, err := conv.Int32(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to int32: %v", err)
		}

		return prepend(INT32, serialize.Int32ToBytes(i)), nil
	case INT64:
		i, err := conv.Int64(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to int64: %v", err)
		}

		return prepend(INT64, serialize.Int64ToBytes(int64(i))), nil
	case BOOLEAN:
		b, err := conv.Bool(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to bool: %v", err)
		}

		return prepend(BOOLEAN, serialize.BoolToBytes(b)), nil
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
	case STRING:
		return serialize.BytesToString(bts), nil
	case INT32:
		return serialize.BytesToInt32(bts), nil
	case INT64:
		return serialize.BytesToInt64(bts), nil
	case BOOLEAN:
		return serialize.BytesToBool(bts), nil
	}

	return nil, fmt.Errorf("unknown type: %d", d)
}
