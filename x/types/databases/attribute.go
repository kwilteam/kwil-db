package databases

import (
	"kwil/x/execution"
	anytype "kwil/x/types/data_types/any_type"
)

type Attribute struct {
	Type  execution.AttributeType `json:"type"`
	Value anytype.Any             `json:"value,omitempty"`
}
