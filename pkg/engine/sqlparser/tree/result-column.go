package tree

import (
	"errors"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"
)

type ResultColumn interface {
	resultColumn()
	ToSQL() string
	Accept(visitor Visitor) error
}

type ResultColumnStar struct{}

func (r *ResultColumnStar) resultColumn() {}
func (r *ResultColumnStar) ToSQL() string {
	return "*"
}
func (r *ResultColumnStar) Accept(visitor Visitor) error {
	return visitor.VisitResultColumnStar(r)
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
func (r *ResultColumnExpression) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitResultColumnExpression(r),
		accept(visitor, r.Expression),
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
func (r *ResultColumnTable) Accept(visitor Visitor) error {
	return visitor.VisitResultColumnTable(r)
}
