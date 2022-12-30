package dto

import "kwil/x/execution"

type Attribute struct {
	Type  execution.AttributeType `json:"type"`
	Value any                     `json:"value,omitempty"`
}
