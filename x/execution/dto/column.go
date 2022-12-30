package dto

import "kwil/x/execution"

type Column struct {
	Name       string             `json:"name"`
	Type       execution.DataType `json:"type"`
	Attributes []*Attribute       `json:"attributes,omitempty"`
}
