package datatypes

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cstockton/go-conv"
)

type dataTypes struct{}

var Utils = &dataTypes{}

// CheckType checks if a string is a valid Kwil Type
func (v *dataTypes) CheckType(s string) error {
	switch s {
	case `null`:
		return nil
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
func (v *dataTypes) StringToKwilType(s string) (DataType, error) {
	s = strings.ToLower(s)

	switch s {
	case `null`:
		return NULL, nil
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
func (c *dataTypes) GolangToKwilType(k reflect.Kind) (DataType, error) {

	switch k {
	case reflect.Invalid:
		return NULL, nil
	case reflect.String:
		return STRING, nil
	case reflect.Int32, reflect.Float32:
		return INT32, nil
	case reflect.Int, reflect.Int64, reflect.Float64:
		return INT64, nil
	case reflect.Bool:
		return BOOLEAN, nil
	}

	return INVALID_DATA_TYPE, fmt.Errorf(`unknown type: "%s"`, k)
}

// Takes the `any` golang type and converts it to a Kwil Type
func (c *dataTypes) AnyToKwilType(val any) (DataType, error) {
	if val == nil {
		return NULL, nil
	}

	valType := reflect.TypeOf(val).Kind()
	return c.GolangToKwilType(valType)
}

// CompareKwilStringToAny compares a Kwil Type to an `any` golang type
func (v *dataTypes) CompareAnyToKwilString(a any, val string) error {
	kwilType, err := Utils.StringToKwilType(val)
	if err != nil {
		return err
	}
	anyType, err := Utils.AnyToKwilType(a)
	if err != nil {
		return err
	}
	if kwilType != anyType {
		return fmt.Errorf(`type mismatch: "%s" != "%s"`, val, a)
	}
	return nil
}

func (c *dataTypes) CompareAnyToKwilType(a any, val DataType) error {
	return c.CompareAnyToKwilString(a, val.String())
}

func (c *dataTypes) KwilToPgType(k DataType) (string, error) {
	switch k {
	case NULL:
		return "", fmt.Errorf(`null type not supported`)
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

func (c *dataTypes) KwilStringToPgType(s string) (string, error) {
	kwilType, err := Utils.StringToKwilType(s)
	if err != nil {
		return ``, err
	}
	return c.KwilToPgType(kwilType)
}

func (c *dataTypes) StringToAnyGolangType(s string, kt DataType) (any, error) {
	switch kt {
	case NULL:
		return nil, nil
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

func (c *dataTypes) ConvertAny(v any, t DataType) (any, error) {
	if v == nil {
		return nil, nil
	}

	switch t {
	case NULL:
		return nil, nil
	case STRING:
		return conv.String(v)
	case INT32:
		return conv.Int32(v)
	case INT64:
		return conv.Int64(v)
	case BOOLEAN:
		return conv.Bool(v)
	default:
		return nil, fmt.Errorf(`unknown type: "%s"`, t.String())
	}
}
