package schema

type UpdateDef struct {
	name    string
	columns ColumnMap
	ifMatch ColumnMap
}

func (q *UpdateDef) Columns() ColumnMap {
	return q.columns
}

func (q *UpdateDef) IfMatch() ColumnMap {
	return q.ifMatch
}

func (q *UpdateDef) Name() string {
	return q.name
}

func (q *UpdateDef) Type() QueryType {
	return Update
}
