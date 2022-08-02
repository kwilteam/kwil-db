package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"
)

type GroupBy struct {
	Expressions []Expression
	Having      Expression
}

func (g *GroupBy) Accept(w Walker) error {
	return run(
		w.EnterGroupBy(g),
		acceptMany(w, g.Expressions),
		accept(w, g.Having),
		w.ExitGroupBy(g),
	)
}

func (g *GroupBy) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(g.Expressions) == 0 {
		panic("no expressions provided to GroupBy")
	}

	stmt.Token.Group().By()

	stmt.WriteList(len(g.Expressions), func(i int) {
		stmt.WriteString(g.Expressions[i].ToSQL())
	})

	if g.Having != nil {
		stmt.Token.Having()
		stmt.WriteString(g.Having.ToSQL())
	}

	return stmt.String()
}
