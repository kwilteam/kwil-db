package dto

type Table struct {
	Name    string    `json:"name"`
	Columns []*Column `json:"columns"`
}

func (t *Table) GetColumn(c string) *Column {
	for _, col := range t.Columns {
		if col.Name == c {
			return col
		}
	}
	return nil
}
