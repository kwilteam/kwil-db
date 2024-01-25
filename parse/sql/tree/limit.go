package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// Limit is a LIMIT clause.
// It takes an expression, and can optionally take either an offset or a second expression.
type Limit struct {
	node

	// LIMIT row_count OFFSET offset;
	// IS SAME AS
	// LIMIT offset, row_count;
	// TODO: in the tree we should just use one or the other, not both.
	Expression       Expression
	Offset           Expression
	SecondExpression Expression
}

func (l *Limit) Accept(v AstVisitor) any {
	return v.VisitLimit(l)
}

// Accept implements the Visitor interface.
func (l *Limit) Walk(w AstWalker) error {
	return run(
		w.EnterLimit(l),
		accept(w, l.Expression),
		accept(w, l.Offset),
		accept(w, l.SecondExpression),
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
