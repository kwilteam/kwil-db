package tree

// Limit is a LIMIT clause.
// It takes an expression, and can optionally take either an offset or a second expression.
type Limit struct {
	Expression       Expression
	Offset           Expression
	SecondExpression Expression
}

// ToSQL marshals a LIMIT clause into a SQL string.
func (l *Limit) ToSQL() string {
	stmt := newSQLBuilder()
	if l.Expression == nil {
		panic("no expression provided to Limit")
	}

	stmt.Write(SPACE, LIMIT, SPACE)
	stmt.WriteString(l.Expression.ToSQL())

	if l.Offset != nil && l.SecondExpression != nil {
		panic("cannot have both offset and second expression in Limit")
	}

	if l.Offset != nil {
		stmt.Write(SPACE, OFFSET, SPACE)
		stmt.WriteString(l.Offset.ToSQL())
	}

	if l.SecondExpression != nil {
		stmt.Write(SPACE, COMMA, SPACE)
		stmt.WriteString(l.SecondExpression.ToSQL())
	}

	return stmt.String()
}
