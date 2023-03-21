package models

import "kwil/pkg/engine/types"

type Attribute struct {
	Type  types.AttributeType `json:"type" clean:"is_enum,attribute_type"`
	Value []byte              `json:"value"`
}
