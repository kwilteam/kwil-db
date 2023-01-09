package validation

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/types/databases"
)

func ValidateColumn(c *databases.Column) error {
	// check if column name is valid
	err := CheckName(c.Name, execution.MAX_COLUMN_NAME_LENGTH)
	if err != nil {
		return fmt.Errorf(`invalid name for column: %w`, err)
	}

	// check attributes amount
	if len(c.Attributes) > execution.MAX_ATTRIBUTES_PER_COLUMN {
		return fmt.Errorf(`column must have at most %d attributes`, execution.MAX_ATTRIBUTES_PER_COLUMN)
	}

	// check if column type is valid
	if !c.Type.IsValid() {
		return fmt.Errorf("unknown column type: %d", c.Type.Int())
	}

	// check if column attributes are valid
	attrMap := make(map[execution.AttributeType]struct{})
	for _, attr := range c.Attributes {
		// check if attribute is unique
		if _, ok := attrMap[attr.Type]; ok {
			return fmt.Errorf(`duplicate attribute "%s" for column "%s"`, attr.Type.String(), c.Name)
		}
		attrMap[attr.Type] = struct{}{}

		// check if attribute is valid
		err = ValidateAttribute(attr, c)
		if err != nil {
			return fmt.Errorf(`invalid attribute for column "%s": %w`, c.Name, err)
		}
	}

	return nil
}
