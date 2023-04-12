package types

import (
	"fmt"

	"github.com/cstockton/go-conv"
)

type DataType int

// Data Types
const (
	INVALID_DATA_TYPE DataType = iota + 100
	NULL
	TEXT
	INT
	END_DATA_TYPE
)

func (d DataType) String() string {
	switch d {
	case NULL:
		return `null`
	case TEXT:
		return `text`
	case INT:
		return `int`
	}
	return `unknown`
}

func (d *DataType) Int() int {
	return int(*d)
}

func (d *DataType) IsNumeric() bool {
	return *d == INT
}

func (d *DataType) IsValid() bool {
	return *d > INVALID_DATA_TYPE && *d < END_DATA_TYPE
}

// will check if the data type is a text type
func (d *DataType) IsText() bool {
	return *d == TEXT
}

// Coerce will try to convert the original value to the data type
func (d DataType) Coerce(original *ConcreteValue) (*ConcreteValue, error) {
	return d.CoerceAnyToConcrete(original.value)
}

// CoerceAny will try to coerce the value to the data type.
// Instead of taking a ConcreteValue, it takes an interface{}.
// This is expected to be scalar values, such as int, string, etc.
func (d DataType) CoerceAny(val any) (any, error) {
	switch d {
	case NULL:
		return nil, nil
	case TEXT:
		return conv.String(val)
	case INT:
		return conv.Int(val)
	}
	return nil, fmt.Errorf("invalid data type for datatype coercion: %d", d.Int())
}

// CoerceAnyToConcrete will try to coerce the value to the data type.
// Instead of taking a ConcreteValue, it takes an interface{}.
// This is expected to be scalar values, such as int, string, etc.
func (d DataType) CoerceAnyToConcrete(val any) (*ConcreteValue, error) {
	switch d {
	case NULL:
		return NewEmpty(), nil
	case TEXT:
		return NewExplicit(val, TEXT)
	case INT:
		return NewExplicit(val, INT)
	}
	return nil, fmt.Errorf("invalid data type for datatype coercion: %d", d.Int())
}
