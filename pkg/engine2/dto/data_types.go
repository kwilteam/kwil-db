package dto

import (
	"fmt"

	"github.com/cstockton/go-conv"
)

type DataType string

// Data Types
const (
	NULL DataType = "NULL"
	TEXT DataType = "TEXT"
	INT  DataType = "INT"
)

func (d DataType) String() string {
	return string(d)
}

func (d *DataType) IsNumeric() bool {
	return *d == INT
}

func (d *DataType) IsValid() bool {
	return *d == NULL || *d == TEXT || *d == INT
}

// will check if the data type is a text type
func (d *DataType) IsText() bool {
	return *d == TEXT
}

// CoerceAny will try to coerce the value to the data type.
// Instead of taking a ConcreteValue, it takes an interface{}.
// This is expected to be scalar values, such as int, string, etc.
func (d DataType) Coerce(val any) (any, error) {
	switch d {
	case NULL:
		return nil, nil
	case TEXT:
		return conv.String(val)
	case INT:
		return conv.Int(val)
	}
	return nil, fmt.Errorf("invalid data type for datatype coercion: %s", d.String())
}
