package clean

import (
	"fmt"
	"kwil/x/execution"
	datatypes "kwil/x/types/data_types"
	"kwil/x/types/databases"
	"kwil/x/utils/serialize"
)

// AssertAttributeType asserts that the attribute type is valid, and converts as necessary.
func AssertAttributeType(attr *databases.Attribute, dataType datatypes.DataType) error {
	// attribute types have required data types for their values
	// the exception is "DEFAULT" which must be whatever type the column is

	// try to convert the attribute value to the correct type
	value, err := serialize.TryUnmarshalType(attr.Value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal attribute value: %w", err)
	}

	switch attr.Type {
	case execution.DEFAULT:
		// convert to column type
		res, err := datatypes.Utils.ConvertAny(value, dataType)
		if err != nil {
			return fmt.Errorf("failed to convert attribute value to column type: %w", err)
		}

		attr.Value = res
	default:
		// convert to attribute type
		res, err := execution.Attributes.ConvertAnyToValidAttributeType(value, attr.Type)
		if err != nil {
			return fmt.Errorf("failed to convert attribute value to attribute type: %w", err)
		}

		attr.Value = res
	}

	return nil
}
