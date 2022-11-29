package schema

type DeleteDef struct {
	name    string
	ifMatch ColumnMap
}

func (q *DeleteDef) IfMatch() ColumnMap {
	return q.ifMatch
}

func (q *DeleteDef) Name() string {
	return q.name
}

func (q *DeleteDef) Type() QueryType {
	return Delete
}
