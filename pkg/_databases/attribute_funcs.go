package databases

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

type attributes struct{}

var Attributes = &attributes{}

func (c *attributes) ValidateAttributeValueType(val spec.DataType, attr spec.AttributeType) error {
	switch attr {
	case spec.PRIMARY_KEY:
		// takes no value
		return nil
	case spec.UNIQUE:
		// takes no value
		return nil
	case spec.NOT_NULL:
		// takes no value
		return nil
	case spec.DEFAULT:
		// can be anything
		return nil
	case spec.MIN:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case spec.MAX:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case spec.MIN_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case spec.MAX_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}

func (c *attributes) DataTypeCanHaveAttribute(attr spec.AttributeType, col spec.DataType) error {
	switch attr {
	case spec.PRIMARY_KEY:
		return nil
	case spec.UNIQUE:
		// can be applied to any type
		return nil
	case spec.NOT_NULL:
		// can be applied to any type
		return nil
	case spec.DEFAULT:
		// can be applied to any type
		return nil
	case spec.MIN:
		if !col.IsNumeric() {
			return fmt.Errorf(`min attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case spec.MAX:
		if !col.IsNumeric() {
			return fmt.Errorf(`max attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case spec.MIN_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`min_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	case spec.MAX_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`max_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}
