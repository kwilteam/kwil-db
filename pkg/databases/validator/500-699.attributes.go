package validator

import (
	"fmt"
	"kwil/pkg/databases"
	datatypes "kwil/pkg/types/data_types"
	anytype "kwil/pkg/types/data_types/any_type"
)

/*
###################################################################################################

	Attributes: 500-599

###################################################################################################
*/

// validateAttributes validates all attributes in an array
// 500 range
func (v *Validator) validateAttributes(col *databases.Column[anytype.KwilAny]) error {
	if len(col.Attributes) > databases.MAX_ATTRIBUTES_PER_COLUMN {
		return violation(errorCode501, fmt.Errorf("too many attributes: %v > %v", len(col.Attributes), databases.MAX_ATTRIBUTES_PER_COLUMN))
	}

	if containsUniqueAndDefaultAttributes(col.Attributes) {
		return violation(errorCode502, fmt.Errorf("column %q contains both UNIQUE and DEFAULT attributes", col.Name))
	}

	attributeNames := make(map[databases.AttributeType]struct{})
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

func containsUniqueAndDefaultAttributes(attributes []*databases.Attribute[anytype.KwilAny]) bool {
	var hasUnique, hasDefault bool
	for _, attr := range attributes {
		if attr.Type == databases.UNIQUE {
			hasUnique = true
		}
		if attr.Type == databases.DEFAULT {
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
func (v *Validator) ValidateAttribute(col *databases.Column[anytype.KwilAny], attr *databases.Attribute[anytype.KwilAny]) error {
	if !attr.Type.IsValid() {
		return violation(errorCode600, fmt.Errorf("unknown attribute type: %v", attr.Type))
	}

	if err := validateAttributeValueType(attr.Value.Type(), attr.Type); err != nil {
		return violation(errorCode601, fmt.Errorf("invalid attribute value type: %w", err))
	}

	if attr.Type == databases.DEFAULT && col.Type != attr.Value.Type() {
		return violation(errorCode602, fmt.Errorf("invalid default value type, default value must be same as column type: %s != %s", col.Type.String(), attr.Value.String()))
	}

	if err := dataTypeCanHaveAttribute(attr.Type, col.Type); err != nil {
		return violation(errorCode603, fmt.Errorf(`column of type %s cannot contain attribute %s: %w`, col.Type.String(), attr.Type.String(), err))
	}

	return nil
}

func dataTypeCanHaveAttribute(attr databases.AttributeType, col datatypes.DataType) error {
	switch attr {
	case databases.PRIMARY_KEY:
		return nil
	case databases.UNIQUE:
		// can be applied to any type
		return nil
	case databases.NOT_NULL:
		// can be applied to any type
		return nil
	case databases.DEFAULT:
		// can be applied to any type
		return nil
	case databases.MIN:
		if !col.IsNumeric() {
			return fmt.Errorf(`min attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case databases.MAX:
		if !col.IsNumeric() {
			return fmt.Errorf(`max attribute can only be applied to numeric types. received: "%s"`, col.String())
		}
		return nil
	case databases.MIN_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`min_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	case databases.MAX_LENGTH:
		if !col.IsText() {
			return fmt.Errorf(`max_length attribute can only be applied to string types. received: "%s"`, col.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}

func validateAttributeValueType(val datatypes.DataType, attr databases.AttributeType) error {
	switch attr {
	case databases.PRIMARY_KEY:
		// takes no value
		return nil
	case databases.UNIQUE:
		// takes no value
		return nil
	case databases.NOT_NULL:
		// takes no value
		return nil
	case databases.DEFAULT:
		// can be anything
		return nil
	case databases.MIN:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case databases.MAX:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case databases.MIN_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`min_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	case databases.MAX_LENGTH:
		// must int
		if !val.IsNumeric() {
			return fmt.Errorf(`max_length attribute must be an int. received: "%s"`, val.String())
		}
		return nil
	}

	return fmt.Errorf(`unknown attribute: "%s"`, attr.String())
}
