package databases

import (
	"fmt"
	"kwil/pkg/types/data_types"
)

type attributes struct{}

var Attributes = &attributes{}

func (c *attributes) ValidateAttributeValueType(val datatypes.DataType, attr AttributeType) error {
	switch attr {
	case PRIMARY_KEY:
		// takes no value
		return nil
	case UNIQUE:
		// takes no value
		return nil
	case NOT_NULL:
		// takes no value
		return nil
	case DEFAULT:
		// can be anything
		return nil
	case MIN:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case MAX:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case MIN_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case MAX_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}

func (c *attributes) DataTypeCanHaveAttribute(attr AttributeType, col datatypes.DataType) error {
	switch attr {
	case PRIMARY_KEY:
		return nil
	case UNIQUE:
		// can be applied to any type
		return nil
	case NOT_NULL:
		// can be applied to any type
		return nil
	case DEFAULT:
		// can be applied to any type
		return nil
	case MIN:
		if !col.IsNumeric() {
			return fmt.Errorf(`min attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case MAX:
		if !col.IsNumeric() {
			return fmt.Errorf(`max attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case MIN_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`min_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	case MAX_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`max_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}
