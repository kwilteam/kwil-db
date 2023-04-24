package models

import (
	"kwil/pkg/engine/types"
)

type Column struct {
	Name       string         `json:"name" clean:"lower"`
	Type       types.DataType `json:"type" clean:"is_enum,data_type"`
	Attributes []*Attribute   `json:"attributes,omitempty" traverse:"shallow"`
}

func (c *Column) GetAttribute(attrType types.AttributeType) *Attribute {
	for _, attr := range c.Attributes {
		if attr.Type == attrType {
			return attr
		}
	}
	return nil
}
