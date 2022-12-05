package schema

type InsertDef struct {
	name    string
	columns ColumnMap
}

func (q *InsertDef) Name() string {
	return q.name
}

func (q *InsertDef) Type() QueryType {
	return Create
}

func (q *InsertDef) Columns() ColumnMap {
	return q.columns
}
