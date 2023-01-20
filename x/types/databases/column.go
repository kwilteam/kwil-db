package databases

import (
	datatypes "kwil/x/types/data_types"
)

type Column[T AnyValue] struct {
	Name       string             `json:"name" clean:"lower"`
	Type       datatypes.DataType `json:"type"`
	Attributes []*Attribute[T]    `json:"attributes,omitempty"`
}
