package anytype

import (
	"fmt"
	datatypes "kwil/x/types/data_types"
	"kwil/x/utils/serialize"

	"github.com/cstockton/go-conv"
)

func marshal(v any, d datatypes.DataType) ([]byte, error) {
	switch d {
	case datatypes.NULL:
		return prepend(datatypes.NULL, nil), nil
	case datatypes.STRING:
		str, err := conv.String(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to string: %v", err)
		}

		return prepend(datatypes.STRING, serialize.StringToBytes(str)), nil
	case datatypes.INT32:
		i, err := conv.Int32(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to int32: %v", err)
		}

		return prepend(datatypes.INT32, serialize.Int32ToBytes(i)), nil
	case datatypes.INT64:
		i, err := conv.Int64(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to int64: %v", err)
		}

		return prepend(datatypes.INT64, serialize.Int64ToBytes(int64(i))), nil
	case datatypes.BOOLEAN:
		b, err := conv.Bool(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert value to bool: %v", err)
		}

		return prepend(datatypes.BOOLEAN, serialize.BoolToBytes(b)), nil
	}

	return nil, fmt.Errorf("unknown type: %d", d)
}

func prepend(b datatypes.DataType, bts []byte) []byte {
	return append([]byte{byte(b)}, bts...)
}

// tryUnmarshal will try to unmarshal a byte slice to the specified type
func tryUnmarshal(bts []byte, d datatypes.DataType) (any, error) {
	switch d {
	case datatypes.NULL:
		return nil, nil
	case datatypes.STRING:
		return serialize.BytesToString(bts), nil
	case datatypes.INT32:
		return serialize.BytesToInt32(bts), nil
	case datatypes.INT64:
		return serialize.BytesToInt64(bts), nil
	case datatypes.BOOLEAN:
		return serialize.BytesToBool(bts), nil
	}

	return nil, fmt.Errorf("unknown type: %d", d)
}
