package validation

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/types/databases"
	"reflect"
)

func ValidateAttribute(a *databases.Attribute, c *databases.Column) error {

	// check if attribute is valid
	if !a.Type.IsValid() {
		return fmt.Errorf(`unknown attribute type: %d`, a.Type.Int())
	}

	// check if attribute value is valid: e.g. if it is a MIN or MAX attribute, the value must be an int
	err := correctAttributeValueType(a.Value, a.Type, c.Type)
	if err != nil {
		return fmt.Errorf(`invalid attribute value: %w`, err)
	}

	// check if attribute is valid for column type
	err = attributeCanBeOnColumnType(a.Type, c.Type)
	if err != nil {
		return fmt.Errorf(`invalid attribute for column type: %w`, err)
	}

	return nil
}

// CorrectAttributeType checks that the attribute value type is correct
// if incorrect, returns an error
func correctAttributeValueType(val any, attr execution.AttributeType, colType execution.DataType) error {
	if val == nil {
		return nil
	}

	valType := reflect.TypeOf(val).Kind()

	switch attr {
	case execution.PRIMARY_KEY:
		return nil
	case execution.UNIQUE:
		return nil
	case execution.NOT_NULL:
		return nil
	case execution.DEFAULT:
		valTypeEnum, err := execution.DataTypes.GolangToKwilType(valType)
		if err != nil {
			return err
		}
		if valTypeEnum != colType {
			return fmt.Errorf(`default value type does not match column type`)
		}
		return nil
	case execution.MIN:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`min attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	case execution.MAX:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`max attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	case execution.MIN_LENGTH:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`min_length attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	case execution.MAX_LENGTH:
		// must int
		if !isNum(valType) {
			return fmt.Errorf(`max_length attribute must be an int. received: "%s"`, valType.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}

// checks if an attribute can be on a column type (for example, we can't have a MIN attribute on a boolean column)
func attributeCanBeOnColumnType(attr execution.AttributeType, colType execution.DataType) error {
	switch attr {
	case execution.PRIMARY_KEY:
		return nil
	case execution.UNIQUE:
		return nil
	case execution.NOT_NULL:
		return nil
	case execution.DEFAULT:
		return nil
	case execution.MIN:
		if !colType.IsNumeric() {
			return fmt.Errorf(`min attribute cannot be on column type "%s"`, colType.String())
		}
		return nil
	case execution.MAX:
		if !colType.IsNumeric() {
			return fmt.Errorf(`max attribute cannot be on column type "%s"`, colType.String())
		}
		return nil
	case execution.MIN_LENGTH:
		if colType != execution.STRING {
			return fmt.Errorf(`min_length attribute cannot be on column type "%s"`, colType.String())
		}
		return nil
	case execution.MAX_LENGTH:
		if colType != execution.STRING {
			return fmt.Errorf(`max_length attribute cannot be on column type "%s"`, colType.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}

// used internally to check if a reflect.Kind is a number
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
