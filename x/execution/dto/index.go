package dto

import "kwil/x/execution"

type Index struct {
	Name    string              `json:"name" yaml:"name"`
	Table   string              `json:"table" yaml:"table"`
	Columns []string            `json:"columns" yaml:"columns"`
	Using   execution.IndexType `json:"using" yaml:"using"`
}
