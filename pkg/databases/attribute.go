package databases

import (
	"kwil/pkg/types/data_types/any_type"
)

type Attribute[T anytype.AnyValue] struct {
	Type  AttributeType `json:"type" clean:"is_enum,attribute_type"`
	Value T             `json:"value,omitempty"`
}
