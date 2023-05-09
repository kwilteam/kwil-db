package databases

import "github.com/kwilteam/kwil-db/pkg/databases/spec"

type Attribute[T spec.AnyValue] struct {
	Type  spec.AttributeType `json:"type" clean:"is_enum,attribute_type"`
	Value T                  `json:"value,omitempty"`
}
