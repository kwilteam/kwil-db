package databases

import (
	"kwil/x/execution"
)

type Attribute[T AnyValue] struct {
	Type  execution.AttributeType `json:"type"`
	Value T                       `json:"value,omitempty"`
}
