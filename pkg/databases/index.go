package databases

import "kwil/pkg/databases/spec"

type Index struct {
	Name    string         `json:"name" clean:"lower"`
	Table   string         `json:"table" clean:"lower"`
	Columns []string       `json:"columns" clean:"lower"`
	Using   spec.IndexType `json:"using" clean:"is_enum,index_type"`
}
