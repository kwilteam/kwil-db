package databases

import (
	datatypes "kwil/x/types/data_types"
	anytype "kwil/x/types/data_types/any_type"
)

type Column[T anytype.AnyValue] struct {
	Name       string             `json:"name" clean:"lower"`
	Type       datatypes.DataType `json:"type" clean:"is_enum,data_type"`
	Attributes []*Attribute[T]    `json:"attributes,omitempty" traverse:"shallow"`
}
