package execution

import (
	"kwil/pkg/types/data_types/any_type"
)

type ExecutionBody[T anytype.AnyValue] struct {
	Database string          `json:"database" yaml:"database" clean:"lower"` // the id
	Query    string          `json:"query" yaml:"query" clean:"lower"`       // the name of the query
	Inputs   []*UserInput[T] `json:"inputs" yaml:"inputs"`                   // the inputs to the query
}
