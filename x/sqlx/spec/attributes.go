package types

import (
	"fmt"
	"reflect"
)

type AttributeType int

// Attributes
const (
	INVALID_ATTRIBUTE AttributeType = iota
	PRIMARY_KEY
	UNIQUE
	NOT_NULL
	DEFAULT
	MIN        // Min allowed value
	MAX        // Max allowed value
	MIN_LENGTH // Min allowed length
	MAX_LENGTH // Max allowed length
)

func (a *AttributeType) String() string {
	switch *a {
	case PRIMARY_KEY:
		return `primary_key`
	case UNIQUE:
		return `unique`
	case NOT_NULL:
		return `not_null`
	case DEFAULT:
		return `default`
	case MIN:
		return `min`
	case MAX:
		return `max`
	case MIN_LENGTH:
		return `min_length`
	case MAX_LENGTH:
		return `max_length`
	}
	return `unknown`
}

// ConvertAttribute converts a string to an AttributeType
func (c *conversion) ConvertAttribute(s string) (AttributeType, error) {
	switch s {
	case `primary_key`:
		return PRIMARY_KEY, nil
	case `unique`:
		return UNIQUE, nil
	case `not_null`:
		return NOT_NULL, nil
	case `default`:
		return DEFAULT, nil
	case `min`:
		return MIN, nil
	case `max`:
		return MAX, nil
	case `min_length`:
		return MIN_LENGTH, nil
	case `max_length`:
		return MAX_LENGTH, nil
	}
	return INVALID_ATTRIBUTE, fmt.Errorf(`unknown attribute: "%s"`, s)
}

func (v *validation) CorrectAttributeType(val any, attr AttributeType, columnType string) error {
	if val == nil {
		return nil
	}

	valType := reflect.TypeOf(val).Kind()
	colType, err := Conversion.StringToKwilType(columnType)
	if err != nil {
		return err
	}

	switch attr {
	case PRIMARY_KEY:
		return nil
	case UNIQUE:
		return nil
	case NOT_NULL:
		return nil
	case DEFAULT:
		valTypeEnum, err := Conversion.GolangToKwilType(valType)
		if err != nil {
			return err
		}
		if valTypeEnum != colType {
			return fmt.Errorf(`default value type does not match column type`)
		}
		return nil
	case MIN:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`min attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	case MAX:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`max attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	case MIN_LENGTH:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`min_length attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	case MAX_LENGTH:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`max_length attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}

func isNum(x reflect.Kind) bool {
	switch x {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	}
	return false
}
