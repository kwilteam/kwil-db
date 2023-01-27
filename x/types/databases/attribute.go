package databases

import (
	"kwil/x/execution"
	anytype "kwil/x/types/data_types/any_type"
)

type Attribute[T anytype.AnyValue] struct {
	Type  execution.AttributeType `json:"type" clean:"is_enum,attribute_type"`
	Value T                       `json:"value,omitempty"`
}
