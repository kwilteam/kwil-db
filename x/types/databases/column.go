package databases

import (
	datatypes "kwil/x/types/data_types"
)

type Column struct {
	Name       string             `json:"name"`
	Type       datatypes.DataType `json:"type"`
	Attributes []*Attribute       `json:"attributes,omitempty"`
}
