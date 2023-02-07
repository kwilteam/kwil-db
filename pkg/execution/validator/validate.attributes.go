package validator

import (
	"fmt"
	execution2 "kwil/pkg/execution"
	"kwil/pkg/types/data_types/any_type"
	databases2 "kwil/pkg/types/databases"
)

/*
	Validate All Attributes
*/

// validateAttributes validates all attributes in an array
func (v *Validator) validateAttributes(col *databases2.Column[anytype.KwilAny]) error {
	// validate attribute count
	err := validateAttributeCount(col.Attributes)
	if err != nil {
		return fmt.Errorf(`invalid attribute count: %w`, err)
	}

	attributeNames := make(map[execution2.AttributeType]struct{})
	for _, attr := range col.Attributes {
		// validate attribute name is unique
		if _, ok := attributeNames[attr.Type]; ok {
			return fmt.Errorf(`duplicate attribute name "%s"`, attr.Type.String())
		}
		attributeNames[attr.Type] = struct{}{}

		// validate attribute
		err := v.validateAttribute(col, attr)
		if err != nil {
			return fmt.Errorf(`error on attribute %s: %w`, attr.Type.String(), err)
		}
	}

	return nil
}

// validateAttributeCount validates the number of attributes in an array
func validateAttributeCount(attributes []*databases2.Attribute[anytype.KwilAny]) error {
	if len(attributes) > databases2.MAX_ATTRIBUTES_PER_COLUMN {
		return fmt.Errorf(`too many attributes: %v > %v`, len(attributes), databases2.MAX_ATTRIBUTES_PER_COLUMN)
	}

	return nil
}

/*
	Validate Attribute
*/

// validateAttribute validates a single attribute
func (v *Validator) validateAttribute(col *databases2.Column[anytype.KwilAny], attr *databases2.Attribute[anytype.KwilAny]) error { // validate attribute type
	err := v.validateAttributeType(attr)
	if err != nil {
		return fmt.Errorf(`invalid attribute type: %w`, err)
	}

	// check that the attribute value type is valid
	err = v.validateAttributeValueType(col, attr)
	if err != nil {
		return fmt.Errorf(`invalid attribute value type: %w`, err)
	}

	err = v.validateColumnCanContainAttributeType(col, attr.Type)
	if err != nil {
		return fmt.Errorf(`invalid attribute type: %w`, err)
	}

	return nil
}

// validateAttributeType validates the type of an attribute
func (v *Validator) validateAttributeType(attr *databases2.Attribute[anytype.KwilAny]) error {
	if !attr.Type.IsValid() {
		return fmt.Errorf(`invalid attribute type: %s`, attr.Type.String())
	}

	return nil
}

func (v *Validator) validateAttributeValueType(col *databases2.Column[anytype.KwilAny], attr *databases2.Attribute[anytype.KwilAny]) error {
	err := execution2.Attributes.ValidateAttributeValueType(attr.Value.Type(), attr.Type)
	if err != nil {
		return fmt.Errorf(`invalid attribute value type: %w`, err)
	}

	if attr.Type == execution2.DEFAULT {
		// check that the default value is valid for the column type
		if col.Type != attr.Value.Type() {
			return fmt.Errorf(`invalid default value type.  default value must be same as column type: %s != %s`, col.Type.String(), attr.Value.String())
		}
	}

	return nil
}

func (v *Validator) validateColumnCanContainAttributeType(col *databases2.Column[anytype.KwilAny], attrType execution2.AttributeType) error {
	err := execution2.Attributes.DataTypeCanHaveAttribute(attrType, col.Type)
	if err != nil {
		return fmt.Errorf(`column of type %s cannot contain attribute %s: %w`, col.Type.String(), attrType.String(), err)
	}

	return nil
}
