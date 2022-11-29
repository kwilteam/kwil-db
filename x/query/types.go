package query

import "kwil/x/schema"

type Insert struct {
	Name    string
	Columns schema.ColumnValues
}

type Delete struct {
	Name    string
	IfMatch schema.ColumnValues
}

type Update struct {
	Name    string
	Columns schema.ColumnValues
	IfMatch schema.ColumnValues
}
