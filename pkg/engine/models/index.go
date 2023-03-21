package models

import "kwil/pkg/engine/types"

type Index struct {
	Name    string          `json:"name" clean:"lower"`
	Columns []string        `json:"columns" clean:"lower"`
	Type    types.IndexType `json:"type" clean:"is_enum,index_type"`
}
