package models

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type attributes struct{}

var Attributes = &attributes{}

func (c *attributes) ValidateAttributeValueType(val types.DataType, attr types.AttributeType) error {
	switch attr {
	case types.PRIMARY_KEY:
		// takes no value
		return nil
	case types.UNIQUE:
		// takes no value
		return nil
	case types.NOT_NULL:
		// takes no value
		return nil
	case types.DEFAULT:
		// can be anything
		return nil
	case types.MIN:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case types.MAX:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case types.MIN_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case types.MAX_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}

func (c *attributes) DataTypeCanHaveAttribute(attr types.AttributeType, col types.DataType) error {
	switch attr {
	case types.PRIMARY_KEY:
		return nil
	case types.UNIQUE:
		// can be applied to any type
		return nil
	case types.NOT_NULL:
		// can be applied to any type
		return nil
	case types.DEFAULT:
		// can be applied to any type
		return nil
	case types.MIN:
		if !col.IsNumeric() {
			return fmt.Errorf(`min attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case types.MAX:
		if !col.IsNumeric() {
			return fmt.Errorf(`max attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case types.MIN_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`min_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	case types.MAX_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`max_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}
