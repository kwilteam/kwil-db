package tree

type GroupBy struct {
	Expressions []Expression
	Having      Expression
}

func (g *GroupBy) ToSQL() string {
	stmt := newSQLBuilder()

	if len(g.Expressions) == 0 {
		panic("no expressions provided to GroupBy")
	}

	stmt.Write(SPACE, GROUP, SPACE, BY, SPACE)
	for i, expr := range g.Expressions {
		if i > 0 && i < len(g.Expressions) {
			stmt.Write(COMMA, SPACE)
		}
		stmt.WriteString(expr.ToSQL())
	}

	if g.Having != nil {
		stmt.Write(SPACE, HAVING, SPACE)
		stmt.WriteString(g.Having.ToSQL())
	}

	return stmt.String()
}
