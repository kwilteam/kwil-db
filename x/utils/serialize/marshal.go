package serialize

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ByteType int

const (
	BYTES ByteType = iota
	STRING
	INT
	INT8
	INT16
	INT32
	INT64
	UINT8
	UINT16
	UINT32
	UINT64
	BOOLEAN
)

// MarshalType converts a value to a byte slice.
// The byte slice is prepended with a byte that represents the type of the value.
func MarshalType(v any) ([]byte, error) {
	if v == nil {
		return []byte{}, nil
	}

	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.String:
		return marshal(v, STRING)
	case reflect.Int:
		return marshal(v, INT)
	case reflect.Int8:
		return marshal(v, INT8)
	case reflect.Int16:
		return marshal(v, INT16)
	case reflect.Int32:
		return marshal(v, INT32)
	case reflect.Int64:
		return marshal(v, INT64)
	case reflect.Uint8:
		return marshal(v, UINT8)
	case reflect.Uint16:
		return marshal(v, UINT16)
	case reflect.Uint32:
		return marshal(v, UINT32)
	case reflect.Uint64:
		return marshal(v, UINT64)
	case reflect.Bool:
		return marshal(v, BOOLEAN)
	}

	return nil, fmt.Errorf("unknown type: %s", kind)
}

// UnmarshalType converts a byte slice to a value.
// The byte slice must be prepended with a byte that represents the type of the value.
func UnmarshalType(bts []byte) (any, error) {
	if len(bts) == 0 {
		return nil, nil
	}

	switch ByteType(bts[0]) {
	case BYTES:
		return bts[1:], nil
	case STRING:
		var v string
		err := unmarshal(bts, &v)
		return v, err
	case INT:
		var v int
		err := unmarshal(bts, &v)
		return v, err
	case INT8:
		var v int8
		err := unmarshal(bts, &v)
		return v, err
	case INT16:
		var v int16
		err := unmarshal(bts, &v)
		return v, err
	case INT32:
		var v int32
		err := unmarshal(bts, &v)
		return v, err
	case INT64:
		var v int64
		err := unmarshal(bts, &v)
		return v, err
	case UINT8:
		var v uint8
		err := unmarshal(bts, &v)
		return v, err
	case UINT16:
		var v uint16
		err := unmarshal(bts, &v)
		return v, err
	case UINT32:
		var v uint32
		err := unmarshal(bts, &v)
		return v, err
	case UINT64:
		var v uint64
		err := unmarshal(bts, &v)
		return v, err
	case BOOLEAN:
		var v bool
		err := unmarshal(bts, &v)
		return v, err
	}

	return nil, fmt.Errorf("unknown type: %d", bts[0])
}

// I include these inner functions since I only want to marshal / unmarshal when I know the type, and I don't want to do it in every switch statement.
func marshal(v any, b ByteType) ([]byte, error) {
	bts, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	bts = append([]byte{byte(b)}, bts...)
	return bts, nil
}

// unmarshal takes bytes, as well as a pointer to a value initialized to the type you want to unmarshal to.
func unmarshal(bts []byte, v any) error {
	err := json.Unmarshal(bts[1:], &v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	return nil
}
