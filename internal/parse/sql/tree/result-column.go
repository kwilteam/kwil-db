package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/internal/parse/sql/tree/sql-writer"
)

// ResultColumnStar represents a wildcard column, i.e. `*`.
type ResultColumnStar struct {
	node
}

func (r *ResultColumnStar) Accept(v AstVisitor) any {
	return v.VisitResultColumnStar(r)
}

func (r *ResultColumnStar) resultColumn() {}
func (r *ResultColumnStar) ToSQL() string {
	return "*"
}
func (r *ResultColumnStar) Walk(w AstListener) error {
	return run(
		w.EnterResultColumnStar(r),
		w.ExitResultColumnStar(r),
	)
}

// ResultColumnExpression represents a result column with an expression and an optional alias.
type ResultColumnExpression struct {
	node

	Expression Expression
	Alias      string
}

func (r *ResultColumnExpression) Accept(v AstVisitor) any {
	return v.VisitResultColumnExpression(r)
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
func (r *ResultColumnExpression) Walk(w AstListener) error {
	return run(
		w.EnterResultColumnExpression(r),
		walk(w, r.Expression),
		w.ExitResultColumnExpression(r),
	)
}

// ResultColumnTable represents a table name with a star, i.e. `table.*`.
// Or qualified star?
type ResultColumnTable struct {
	node

	TableName string
}

func (r *ResultColumnTable) Accept(v AstVisitor) any {
	return v.VisitResultColumnTable(r)
}

func (r *ResultColumnTable) resultColumn() {}
func (r *ResultColumnTable) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteIdent(r.TableName)
	stmt.Token.Period()
	stmt.Token.Asterisk()
	return stmt.String()
}
func (r *ResultColumnTable) Walk(w AstListener) error {
	return run(
		w.EnterResultColumnTable(r),
		w.ExitResultColumnTable(r),
	)
}
