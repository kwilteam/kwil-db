package tree

import (
	"errors"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"
)

// Limit is a LIMIT clause.
// It takes an expression, and can optionally take either an offset or a second expression.
type Limit struct {
	Expression       Expression
	Offset           Expression
	SecondExpression Expression
}

// Accept implements the Visitor interface.
func (l *Limit) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitLimit(l),
		accept(visitor, l.Expression),
		accept(visitor, l.Offset),
		accept(visitor, l.SecondExpression),
	)
}

// ToSQL marshals a LIMIT clause into a SQL string.
func (l *Limit) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	if l.Expression == nil {
		panic("no expression provided to Limit")
	}

	stmt.Token.Limit()
	stmt.WriteString(l.Expression.ToSQL())

	if l.Offset != nil && l.SecondExpression != nil {
		panic("cannot have both offset and second expression in Limit")
	}

	if l.Offset != nil {
		stmt.Token.Offset()
		stmt.WriteString(l.Offset.ToSQL())
	}

	if l.SecondExpression != nil {
		stmt.Token.Comma()
		stmt.WriteString(l.SecondExpression.ToSQL())
	}

	return stmt.String()
}
