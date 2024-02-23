package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// Limit is a LIMIT clause.
// It takes an expression, and can optionally take either an offset or a second expression.
type Limit struct {
	Expression Expression
	Offset     Expression
}

// Accept implements the Visitor interface.
func (l *Limit) Accept(w Walker) error {
	return run(
		w.EnterLimit(l),
		accept(w, l.Expression),
		accept(w, l.Offset),
		w.ExitLimit(l),
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

	if l.Offset != nil {
		stmt.Token.Offset()
		stmt.WriteString(l.Offset.ToSQL())
	}

	return stmt.String()
}
