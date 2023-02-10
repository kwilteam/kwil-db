package spec

import (
	"fmt"
	"reflect"
	"strings"
)

type dataTypes struct{}

var DataTypeConversions = &dataTypes{}

// String to Kwil Type converts a string received from JSON/YAML to a Kwil Type
func (c *dataTypes) StringToKwilType(s string) (DataType, error) {
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
