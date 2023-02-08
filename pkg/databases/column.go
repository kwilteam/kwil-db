package databases

import (
	datatypes "kwil/pkg/types/data_types"
	anytype "kwil/pkg/types/data_types/any_type"
)

type Column[T anytype.AnyValue] struct {
	Name       string             `json:"name" clean:"lower"`
	Type       datatypes.DataType `json:"type" clean:"is_enum,data_type"`
	Attributes []*Attribute[T]    `json:"attributes,omitempty" traverse:"shallow"`
}

func (c *Column[T]) GetAttribute(attrType AttributeType) *Attribute[T] {
	for _, attr := range c.Attributes {
		if attr.Type == attrType {
			return attr
		}
	}
	return nil
}
