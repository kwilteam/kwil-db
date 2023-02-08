package spec

import (
	b64 "encoding/base64"
	"fmt"

	"github.com/cstockton/go-conv"
)

// the AnyValue is used to specify what the "any" type of value stored is.
// for example, attributes can hold an "any" type, but sometimes we need this as a string,
// and sometimes we need it as an anytype.KwilAny, which allows us to convert it to
// a Kwil native type.
type AnyValue interface {
	*KwilAny | []byte
}

// The custom KwilAny type is used to store an untyped value of any data type supported by Kwil.
// It includes methods to serialize and deserialize the value back to its original type.
type KwilAny struct {
	value    any
	bytes    []byte
	dataType DataType
}

// New creates a corresponding value type for the given value.
func New(v any) (*KwilAny, error) {
	dataType, err := DataTypeConversions.AnyToKwilType(v)
	if err != nil {
		return nil, fmt.Errorf("failed to get kwil type from value: %w", err)
	}

	return newAny(v, dataType)
}

func NewExplicit(v any, dataType DataType) (*KwilAny, error) {
	return newAny(v, dataType)
}

func newAny(v any, dataType DataType) (*KwilAny, error) {
	// marshal the value
	var bts []byte
	if v != nil {
		var err error
		bts, err = marshal(v, dataType)
		if err != nil {
			return &KwilAny{}, fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	return &KwilAny{
		value:    v,
		bytes:    bts,
		dataType: dataType,
	}, nil
}

// NewMust is like New but panics if an error occurs.
func NewMust(v any) *KwilAny {
	a, err := New(v)
	if err != nil {
		panic(err)
	}
	return a
}

// NewFromSerial creates a corresponding value type for the given serialized value.
func NewFromSerial(b []byte) (*KwilAny, error) {
	if len(b) == 0 {
		return New(nil)
	}

	dataType := DataType(b[0])
	if !dataType.IsValid() {
		return nil, fmt.Errorf("serialized value starts with invalid data type identifier: %v", dataType)
	}

	// try to unmarshal the value
	value, err := tryUnmarshal(b[1:], dataType)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal serialized value: %w", err)
	}

	return &KwilAny{
		value:    value,
		dataType: dataType,
		bytes:    b,
	}, nil
}

// Bytes returns the serialized value.
func (a *KwilAny) Bytes() []byte {
	return a.bytes
}

// Value returns the native value.
func (a *KwilAny) Value() any {
	return a.value
}

// GetType returns the native data type of the value.
func (a *KwilAny) Type() DataType {
	return a.dataType
}

// Copy returns a copy of the struct.
func (a *KwilAny) Copy() KwilAny {
	return KwilAny{
		bytes:    a.bytes,
		value:    a.value,
		dataType: a.dataType,
	}
}

func (a *KwilAny) Base64() string {
	return b64.StdEncoding.EncodeToString(a.bytes)
}

// String returns the value deserialized and converted to a string.
func (a *KwilAny) String() string {
	return fmt.Sprintf("%v", a.value)
}

// IsEmpty returns true if the value is nil.  For example, if there is a string value
// of "", this will return false.  0 is also nil for int types.
func (a *KwilAny) IsEmpty() bool {
	switch a.dataType {
	case STRING:
		return a.value == nil || a.value.(string) == ""
	case INT32:
		return a.value == nil || a.value.(int32) == 0
	case INT64:
		return a.value == nil || a.value.(int64) == 0
	case BOOLEAN:
		return a.value == nil || !a.value.(bool)
	}

	return a.value == nil
}

func (a *KwilAny) Print() {
	fmt.Println("Value:   ", a.value)
	fmt.Println("Bytes:   ", a.bytes)
	fmt.Println("Type:    ", a.dataType.String())
	fmt.Println("Encoded: ", a.Base64())
}

func (a *KwilAny) AsInt() (int, error) {
	return conv.Int(a.value)
}

func (a *KwilAny) AsInt32() (int32, error) {
	return conv.Int32(a.value)
}

func (a *KwilAny) AsInt64() (int64, error) {
	return conv.Int64(a.value)
}

func (a *KwilAny) AsString() (string, error) {
	return conv.String(a.value)
}

func (a *KwilAny) AsBool() (bool, error) {
	return conv.Bool(a.value)
}
