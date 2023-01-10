package spec

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cstockton/go-conv"
)

type DataType int

// Data Types
const (
	INVALID_DATA_TYPE DataType = iota
	STRING
	INT32
	INT64
	BOOLEAN
)

func (d *DataType) String() string {
	switch *d {
	case STRING:
		return `string`
	case INT32:
		return `int32`
	case INT64:
		return `int64`
	case BOOLEAN:
		return `boolean`
	}
	return `unknown`
}

func (d *DataType) Int() int {
	return int(*d)
}

func (v *validation) CheckType(s string) error {
	switch s {
	case `string`:
		return nil
	case `int32`:
		return nil
	case `int64`:
		return nil
	case `boolean`:
		return nil
	}
	return fmt.Errorf(`unknown type: "%s"`, s)
}

// String to Kwil Type converts a string received from JSON/YAML to a Kwil Type
func (v *conversion) StringToKwilType(s string) (DataType, error) {
	s = strings.ToLower(s)

	switch s {
	case `string`:
		return STRING, nil
	case `int32`:
		return INT32, nil
	case `int64`:
		return INT64, nil
	case `boolean`:
		return BOOLEAN, nil
	}
	return INVALID_DATA_TYPE, fmt.Errorf(`unknown type: "%s"`, s)
}

// Golang to Kwil Type converts a reflect.Kind to a Kwil Type
func (c *conversion) GolangToKwilType(k reflect.Kind) (DataType, error) {
	switch k {
	case reflect.String:
		return STRING, nil
	case reflect.Int32 | reflect.Float32:
		return INT32, nil
	case reflect.Int64 | reflect.Float64:
		return INT64, nil
	case reflect.Bool:
		return BOOLEAN, nil
	}

	return INVALID_DATA_TYPE, fmt.Errorf(`unknown type: "%s"`, k)
}

// Takes the `any` golang type and converts it to a Kwil Type
func (c *conversion) AnyToKwilType(val any) (DataType, error) {
	if val == nil {
		return INVALID_DATA_TYPE, fmt.Errorf(`cannot convert nil to Kwil Type`)
	}

	valType := reflect.TypeOf(val).Kind()
	return c.GolangToKwilType(valType)
}

// CompareKwilStringToAny compares a Kwil Type to an `any` golang type
func (v *validation) CompareAnyToKwilString(a any, val string) error {
	kwilType, err := Conversion.StringToKwilType(val)
	if err != nil {
		return err
	}
	anyType, err := Conversion.AnyToKwilType(a)
	if err != nil {
		return err
	}
	if kwilType != anyType {
		return fmt.Errorf(`type mismatch: "%s" != "%s"`, val, a)
	}
	return nil
}

func (c *conversion) KwilToPgType(k DataType) (string, error) {
	switch k {
	case STRING:
		return `text`, nil
	case INT32:
		return `int4`, nil
	case INT64:
		return `int8`, nil
	case BOOLEAN:
		return `boolean`, nil
	}
	return ``, fmt.Errorf(`unknown type: "%s"`, k.String())
}

func (c *conversion) KwilStringToPgType(s string) (string, error) {
	kwilType, err := Conversion.StringToKwilType(s)
	if err != nil {
		return ``, err
	}
	return c.KwilToPgType(kwilType)
}

func (c *conversion) StringToAnyGolangType(s string, kt DataType) (any, error) {
	switch kt {
	case STRING:
		return s, nil
	case INT32:
		return conv.Int32(s)
	case INT64:
		return conv.Int64(s)
	case BOOLEAN:
		return conv.Bool(s)
	default:
		return nil, fmt.Errorf(`unknown type: "%s"`, s)
	}
}
