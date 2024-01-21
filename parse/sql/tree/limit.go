package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// Limit is a LIMIT clause.
// It takes an expression, and can optionally take either an offset or a second expression.
type Limit struct {
	node

	Expression Expression
	Offset     Expression
}

func (l *Limit) Accept(v AstVisitor) any {
	return v.VisitLimit(l)
}

// Accept implements the Visitor interface.
func (l *Limit) Walk(w AstListener) error {
	return run(
		w.EnterLimit(l),
		walk(w, l.Expression),
		walk(w, l.Offset),
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
