package databases

import "kwil/pkg/databases/spec"

type Table[T spec.AnyValue] struct {
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
