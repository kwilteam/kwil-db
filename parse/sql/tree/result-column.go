package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type ResultColumn interface {
	resultColumn()
	ToSQL() string
	Accept(w Walker) error
}

type ResultColumnStar struct{}

func (r *ResultColumnStar) resultColumn() {}
func (r *ResultColumnStar) ToSQL() string {
	return "*"
}
func (r *ResultColumnStar) Accept(w Walker) error {
	return run(
		w.EnterResultColumnStar(r),
		w.ExitResultColumnStar(r),
	)
}

type ResultColumnExpression struct {
	Expression Expression
	Alias      string
}

func (r *ResultColumnExpression) resultColumn() {}
func (r *ResultColumnExpression) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteString(r.Expression.ToSQL())
	if r.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdent(r.Alias)
	}
	return stmt.String()
}
func (r *ResultColumnExpression) Accept(w Walker) error {
	return run(
		w.EnterResultColumnExpression(r),
		accept(w, r.Expression),
		w.ExitResultColumnExpression(r),
	)
}

type ResultColumnTable struct {
	TableName string
}

func (r *ResultColumnTable) resultColumn() {}
func (r *ResultColumnTable) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteIdent(r.TableName)
	stmt.Token.Period()
	stmt.Token.Asterisk()
	return stmt.String()
}
func (r *ResultColumnTable) Accept(w Walker) error {
	return run(
		w.EnterResultColumnTable(r),
		w.ExitResultColumnTable(r),
	)
}
