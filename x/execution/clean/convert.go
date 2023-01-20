package clean

import (
	"fmt"
	"kwil/x/execution"
	datatypes "kwil/x/types/data_types"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

// AssertAttributeType asserts that the attribute type is valid, and converts as necessary.
func AssertAttributeType(attr *databases.Attribute[anytype.KwilAny], dataType datatypes.DataType) error {
	// attribute types have required data types for their values
	// the exception is "DEFAULT" which must be whatever type the column is

	switch attr.Type {
	case execution.DEFAULT:
		// convert to column type
		res, err := datatypes.Utils.ConvertAny(attr.Value, dataType)
		if err != nil {
			return fmt.Errorf("failed to convert attribute value to column type: %w", err)
		}

		attr.Value = res
	default:
		// convert to attribute type
		res, err := execution.Attributes.ConvertAnyToValidAttributeType(attr.Value, attr.Type)
		if err != nil {
			return fmt.Errorf("failed to convert attribute value to attribute type: %w", err)
		}

		attr.Value = res
	}

	return nil
}
