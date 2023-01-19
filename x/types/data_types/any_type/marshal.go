package anytype

import (
	"encoding/json"
	"fmt"
	datatypes "kwil/x/types/data_types"
)

// I include these inner functions since I only want to marshal / unmarshal when I know the type, and I don't want to do it in every switch statement.
func marshal(v any, d datatypes.DataType) ([]byte, error) {
	bts, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	bts = append([]byte{byte(d)}, bts...)
	return bts, nil
}

// tryUnmarshal will try to unmarshal a byte slice to the specified type
func tryUnmarshal(bts []byte, d datatypes.DataType) (any, error) {
	switch d {
	case datatypes.NULL:
		return nil, nil
	case datatypes.STRING:
		var v string
		err := unmarshal(bts, &v)
		return v, err
	case datatypes.INT32:
		var v int32
		err := unmarshal(bts, &v)
		return v, err
	case datatypes.INT64:
		var v int64
		err := unmarshal(bts, &v)
		return v, err
	case datatypes.BOOLEAN:
		var v bool
		err := unmarshal(bts, &v)
		return v, err
	}

	return nil, fmt.Errorf("unknown type: %d", d)
}

// unmarshal takes bytes, as well as a pointer to a value initialized to the type you want to unmarshal to.
func unmarshal(bts []byte, v any) error {
	err := json.Unmarshal(bts, &v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	return nil
}
