package types

import (
	b64 "encoding/base64"
	"fmt"

	"github.com/cstockton/go-conv"
)

// The custom ConcreteValue type is used to store an untyped value of any data type supported by Kwil.
// It includes methods to serialize and deserialize the value back to its original type.
type ConcreteValue struct {
	value    any
	bytes    []byte
	dataType DataType
}

// New creates a corresponding value type for the given value.
func New(v any) (*ConcreteValue, error) {
	dataType, err := DataTypeConversions.AnyToKwilType(v)
	if err != nil {
		return nil, fmt.Errorf("failed to get kwil type from value: %w", err)
	}

	return newAny(v, dataType)
}

func NewExplicit(v any, dataType DataType) (*ConcreteValue, error) {
	return newAny(v, dataType)
}

func NewEmpty() *ConcreteValue {
	bts, err := marshal(nil, NULL)
	if err != nil {
		panic(err)
	}

	return &ConcreteValue{
		value:    nil,
		bytes:    bts,
		dataType: NULL,
	}
}

func newAny(v any, dataType DataType) (*ConcreteValue, error) {
	// marshal the value
	var bts []byte
	if v != nil {
		var err error
		bts, err = marshal(v, dataType)
		if err != nil {
			return &ConcreteValue{}, fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	return &ConcreteValue{
		value:    v,
		bytes:    bts,
		dataType: dataType,
	}, nil
}

// NewMust is like New but panics if an error occurs.
func NewMust(v any) *ConcreteValue {
	a, err := New(v)
	if err != nil {
		panic(err)
	}
	return a
}

// NewFromSerial creates a corresponding value type for the given serialized value.
func NewFromSerial(b []byte) (*ConcreteValue, error) {
	if len(b) == 0 {
		return New(nil)
	}

	dataType := DataType(b[0])
	if !dataType.IsValid() {
		return nil, fmt.Errorf("serialized value starts with invalid data type identifier: %v", dataType.Int())
	}

	// try to unmarshal the value
	value, err := tryUnmarshal(b[1:], dataType)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal serialized value: %w", err)
	}

	return &ConcreteValue{
		value:    value,
		dataType: dataType,
		bytes:    b,
	}, nil
}

// Bytes returns the serialized value.
func (a *ConcreteValue) Bytes() []byte {
	return a.bytes
}

// Value returns the native value.
func (a *ConcreteValue) Value() any {
	return a.value
}

// GetType returns the native data type of the value.
func (a *ConcreteValue) Type() DataType {
	return a.dataType
}

// Copy returns a copy of the struct.
func (a *ConcreteValue) Copy() ConcreteValue {
	return ConcreteValue{
		bytes:    a.bytes,
		value:    a.value,
		dataType: a.dataType,
	}
}

func (a *ConcreteValue) Base64() string {
	return b64.StdEncoding.EncodeToString(a.bytes)
}

// String returns the value deserialized and converted to a string.
func (a *ConcreteValue) String() string {
	return fmt.Sprintf("%v", a.value)
}

// IsEmpty returns true if the value is nil.  For example, if there is a string value
// of "", this will return false.  0 is also nil for int types.
func (a *ConcreteValue) IsEmpty() bool {
	switch a.dataType {
	case TEXT:
		return a.value == nil || a.value.(string) == ""
	case INT:
		return a.value == nil || a.value.(int) == 0
	}

	return a.value == nil
}

func (a *ConcreteValue) Print() {
	fmt.Println("Value:   ", a.value)
	fmt.Println("Bytes:   ", a.bytes)
	fmt.Println("Type:    ", a.dataType.String())
	fmt.Println("Encoded: ", a.Base64())
}

func (a *ConcreteValue) AsInt() (int, error) {
	return conv.Int(a.value)
}

func (a *ConcreteValue) AsString() (string, error) {
	return conv.String(a.value)
}

func (a *ConcreteValue) AsBool() (bool, error) {
	return conv.Bool(a.value)
}
