package dml

import "github.com/doug-martin/goqu/v9"

func BuildInsert(schemaName, table string, columns []any) (string, error) {
	tblName := makeTableName(schemaName, table)
	stmt, _, err := goqu.Dialect("postgres").Insert(tblName).Prepared(true).Cols(columns...).Vals(columns).ToSQL() // anything can be passed to values since it is prepared, but it must be the same length as columns
	return stmt, err
}
