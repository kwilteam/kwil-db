package databases

import (
	"kwil/pkg/types/data_types/any_type"
)

type Table[T anytype.AnyValue] struct {
	Name    string       `json:"name" clean:"lower"`
	Columns []*Column[T] `json:"columns"`
}

func (t *Table[T]) GetColumn(c string) *Column[T] {
	for _, col := range t.Columns {
		if col.Name == c {
			return col
		}
	}
	return nil
}
