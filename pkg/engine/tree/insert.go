package tree

import (
	"fmt"
)

type InsertStatement struct {
	stmt            *insertBuilder
	CTEs            []*CTE
	Table           string
	TableAlias      string
	Columns         []string
	Values          [][]InsertExpression
	Upsert          *Upsert
	ReturningClause *ReturningClause
}

func (i *InsertStatement) ToSql() (string, []any, error) {
	if i.Table == "" {
		return "", nil, fmt.Errorf("sql syntax error: insert does not contain table name")
	}

	i.stmt = Builder.InsertInto(i.Table)
	if i.TableAlias != "" {
		i.stmt = i.stmt.As(i.TableAlias)
	}
	if len(i.Columns) > 0 {
		i.stmt = i.stmt.Columns(i.columns())
	}

	if len(i.Values) > 0 {
		for _, exprs := range i.Values {
			i.stmt = i.stmt.Values(exprs...)
		}
	}

	if i.Upsert != nil {
		i.stmt = i.stmt.WithUpsert(i.Upsert)
	}

	return i.stmt.ToSql()
}

func (i *InsertStatement) columns() []any {
	cols := make([]any, len(i.Columns))
	for _, col := range i.Columns {
		cols = append(cols, col)
	}
	return cols
}
