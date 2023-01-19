package anytype

import (
	"fmt"
	"kwil/x/logx"
	datatypes "kwil/x/types/data_types"
)

// The custom Any type is used to store a value of any type.
// It includes methods to serialize and unserialize the value back to its original type.
type Any struct {
	Serialized bool               `json:"serialized"`
	Value      any                `json:"value"`
	Type       datatypes.DataType `json:"type"`
}

// New creates a corresponding Any type for the given value.
func New(v any) (*Any, error) {
	dataType, err := datatypes.Utils.AnyToKwilType(v)
	if err != nil {
		return nil, fmt.Errorf("failed to get kwil type from any: %w", err)
	}

	return &Any{
		Serialized: false,
		Value:      v,
		Type:       dataType,
	}, nil
}

// NewFromSerial creates a corresponding Any type for the given serialized value.
func NewFromSerial(v []byte) (*Any, error) {
	if len(v) == 0 {
		return New(nil)
	}

	dataType := datatypes.DataType(v[0])
	if dataType <= datatypes.INVALID_DATA_TYPE || dataType >= datatypes.END_DATA_TYPE {
		return nil, fmt.Errorf("serialized value starts with invalid data type identifier: %v", dataType)
	}

	return &Any{
		Serialized: true,
		Value:      v[1:],
		Type:       dataType,
	}, nil
}

// Unserialize unserializes the value if it is serialized.
// It will return the unserialized value.
func (a *Any) Unserialize() (any, error) {
	value, err := a.GetUnserialized()
	if err != nil {
		return nil, fmt.Errorf("failed to get unserialized value: %w", err)
	}

	a.Value = value
	a.Serialized = false

	return value, nil
}

// Serialize serializes the value if it is not serialized.
// It will return the serialized value.
func (a *Any) Serialize() ([]byte, error) {
	value, err := a.GetSerialized()
	if err != nil {
		return nil, fmt.Errorf("failed to get serialized value: %w", err)
	}

	a.Value = value
	a.Serialized = true

	return value, nil
}

// GetSerialized copies the serialized value.
// It does NOT change the state of the Any type.
func (a *Any) GetSerialized() ([]byte, error) {
	if a.Value == nil {
		return nil, nil
	}

	if a.Serialized {
		val, ok := a.Value.([]byte)
		if !ok {
			logx.New().Error("serialized value is not a byte slice.  this should never happen")
			return nil, fmt.Errorf("serialized value is not a byte slice")
		}
		return val, nil
	}

	return marshal(a.Value, a.Type)
}

// GetUnserialized copies the unserialized value.
// It does NOT change the state of the Any type.
func (a *Any) GetUnserialized() (any, error) {
	if a.Value == nil {
		return nil, nil
	}

	if !a.Serialized {
		return a.Value, nil
	}

	bts, ok := a.Value.([]byte)
	if !ok {
		logx.New().Error("serialized value is not a byte slice")
	}

	return tryUnmarshal(bts, a.Type)
}
