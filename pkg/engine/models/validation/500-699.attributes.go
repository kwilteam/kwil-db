package validation

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

/*
###################################################################################################

	Attributes: 500-599

###################################################################################################
*/

// validateAttributes validates all attributes in an array
// 500 range
func validateAttributes(col *models.Column) error {
	if len(col.Attributes) > MAX_ATTRIBUTES_PER_COLUMN {
		return violation(errorCode501, fmt.Errorf("too many attributes: %v > %v", len(col.Attributes), MAX_ATTRIBUTES_PER_COLUMN))
	}

	if containsUniqueAndDefaultAttributes(col.Attributes) {
		return violation(errorCode502, fmt.Errorf("column %q contains both UNIQUE and DEFAULT attributes", col.Name))
	}

	attributeNames := make(map[types.AttributeType]struct{})
	for _, attr := range col.Attributes {
		if _, ok := attributeNames[attr.Type]; ok {
			return violation(errorCode500, fmt.Errorf("duplicate attribute type %v", attr.Type))
		}
		attributeNames[attr.Type] = struct{}{}

		if err := ValidateAttribute(col, attr); err != nil {
			return fmt.Errorf("error on attribute %s: %w", attr.Type.String(), err)
		}
	}

	return nil
}

func containsUniqueAndDefaultAttributes(attributes []*models.Attribute) bool {
	var hasUnique, hasDefault bool
	for _, attr := range attributes {
		if attr.Type == types.UNIQUE {
			hasUnique = true
		}
		if attr.Type == types.DEFAULT {
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
func ValidateAttribute(col *models.Column, attr *models.Attribute) error {
	attributeVal, err := types.NewFromSerial(attr.Value)
	if err != nil {
		return violation(errorCode600, fmt.Errorf("invalid attribute value: %w", err))
	}

	if !attr.Type.IsValid() {
		return violation(errorCode600, fmt.Errorf("unknown attribute type: %v", attr.Type))
	}

	if err := validateAttributeValueType(attributeVal.Type(), attr.Type); err != nil {
		return violation(errorCode601, fmt.Errorf("invalid attribute value type: %w", err))
	}

	if attr.Type == types.DEFAULT && col.Type != attributeVal.Type() {
		return violation(errorCode602, fmt.Errorf("invalid default value type, default value must be same as column type: %s != %s", col.Type.String(), attributeVal.String()))
	}

	if err := dataTypeCanHaveAttribute(attr.Type, col.Type); err != nil {
		return violation(errorCode603, fmt.Errorf(`column of type %s cannot contain attribute %s: %w`, col.Type.String(), attr.Type.String(), err))
	}

	return nil
}

func dataTypeCanHaveAttribute(attr types.AttributeType, col types.DataType) error {
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

func validateAttributeValueType(val types.DataType, attr types.AttributeType) error {
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
