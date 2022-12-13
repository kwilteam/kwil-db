package models

import (
	"fmt"
	types "kwil/x/sqlx"
)

type Column struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Attributes []*Attribute `json:"attributes,omitempty"`
}

func (c *Column) Validate() error {
	// check if column name is valid
	err := CheckName(c.Name, types.COLUMN)
	if err != nil {
		return fmt.Errorf(`invalid name for column: %w`, err)
	}

	// check if column type is valid
	err = types.Validation.CheckType(c.Type)
	if err != nil {
		return fmt.Errorf(`invalid type for column "%s": %w`, c.Name, err)
	}

	// check if column attributes are valid
	attrMap := make(map[string]struct{})
	for _, attr := range c.Attributes {
		// check if attribute is unique
		if _, ok := attrMap[attr.Type]; ok {
			return fmt.Errorf(`duplicate attribute "%s" for column "%s"`, attr.Type, c.Name)
		}
		attrMap[attr.Type] = struct{}{}

		// check if attribute is valid
		err = attr.Validate(c)
		if err != nil {
			return fmt.Errorf(`invalid attribute for column "%s": %w`, c.Name, err)
		}
	}

	return nil
}

func (c *Column) GetName() string {
	return c.Name
}
