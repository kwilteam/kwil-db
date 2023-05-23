package models

type Table struct {
	Name    string    `json:"name" clean:"lower"`
	Columns []*Column `json:"columns"`
	Indexes []*Index  `json:"indexes"`
}

func (t *Table) GetColumn(c string) *Column {
	for _, col := range t.Columns {
		if col.Name == c {
			return col
		}
	}
	return nil
}

func (t *Table) ListColumns() []string {
	var columns []string
	for _, col := range t.Columns {
		columns = append(columns, col.Name)
	}
	return columns
}
