package tree

import "github.com/doug-martin/goqu/v9/exp"

type SelectStatement struct {
	CTEs       []*CTE
	SelectType SelectType
	Columns    []string
	From       *FromClause
}

type SelectType uint8

const (
	SelectTypeAll SelectType = iota
	SelectTypeDistinct
)

type FromClause struct {
	TableOrSubquery *TableOrSubquery
	JoinClauses     []*JoinClause
}

type WhereClause struct {
	Expression Expression
}

func (w *WhereClause) ToSqlStruct() any {
	return w.Expression.ToSqlStruct()
}

func (w *WhereClause) toGoquExpr() exp.Expression {
	return w.Expression.ToSqlStruct().(exp.Expression)
}
