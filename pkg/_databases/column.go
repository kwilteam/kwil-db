package databases

import (
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

type Column[T spec.AnyValue] struct {
	Name       string          `json:"name" clean:"lower"`
	Type       spec.DataType   `json:"type" clean:"is_enum,data_type"`
	Attributes []*Attribute[T] `json:"attributes,omitempty" traverse:"shallow"`
}

func (c *Column[T]) GetAttribute(attrType spec.AttributeType) *Attribute[T] {
	for _, attr := range c.Attributes {
		if attr.Type == attrType {
			return attr
		}
	}
	return nil
}
