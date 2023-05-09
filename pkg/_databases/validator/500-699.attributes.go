package validator

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

/*
###################################################################################################

	Attributes: 500-599

###################################################################################################
*/

// validateAttributes validates all attributes in an array
// 500 range
func (v *Validator) validateAttributes(col *databases.Column[*spec.KwilAny]) error {
	if len(col.Attributes) > MAX_ATTRIBUTES_PER_COLUMN {
		return violation(errorCode501, fmt.Errorf("too many attributes: %v > %v", len(col.Attributes), MAX_ATTRIBUTES_PER_COLUMN))
	}

	if containsUniqueAndDefaultAttributes(col.Attributes) {
		return violation(errorCode502, fmt.Errorf("column %q contains both UNIQUE and DEFAULT attributes", col.Name))
	}

	attributeNames := make(map[spec.AttributeType]struct{})
	for _, attr := range col.Attributes {
		if _, ok := attributeNames[attr.Type]; ok {
			return violation(errorCode500, fmt.Errorf("duplicate attribute type %v", attr.Type))
		}
		attributeNames[attr.Type] = struct{}{}

		if err := v.ValidateAttribute(col, attr); err != nil {
			return fmt.Errorf("error on attribute %d: %w", attr.Type, err)
		}
	}

	return nil
}

func containsUniqueAndDefaultAttributes(attributes []*databases.Attribute[*spec.KwilAny]) bool {
	var hasUnique, hasDefault bool
	for _, attr := range attributes {
		if attr.Type == spec.UNIQUE {
			hasUnique = true
		}
		if attr.Type == spec.DEFAULT {
			hasDefault = true
		}
	}
	return hasUnique && hasDefault
}

/*
###################################################################################################

	Attribute: 600-699

###################################################################################################
*/

// ValidateAttribute validates a single attribute
func (v *Validator) ValidateAttribute(col *databases.Column[*spec.KwilAny], attr *databases.Attribute[*spec.KwilAny]) error {
	if !attr.Type.IsValid() {
		return violation(errorCode600, fmt.Errorf("unknown attribute type: %v", attr.Type))
	}

	if err := validateAttributeValueType(attr.Value.Type(), attr.Type); err != nil {
		return violation(errorCode601, fmt.Errorf("invalid attribute value type: %w", err))
	}

	if attr.Type == spec.DEFAULT && col.Type != attr.Value.Type() {
		return violation(errorCode602, fmt.Errorf("invalid default value type, default value must be same as column type: %s != %s", col.Type.String(), attr.Value.String()))
	}

	if err := dataTypeCanHaveAttribute(attr.Type, col.Type); err != nil {
		return violation(errorCode603, fmt.Errorf(`column of type %s cannot contain attribute %s: %w`, col.Type.String(), attr.Type.String(), err))
	}

	return nil
}

func dataTypeCanHaveAttribute(attr spec.AttributeType, col spec.DataType) error {
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

func validateAttributeValueType(val spec.DataType, attr spec.AttributeType) error {
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
