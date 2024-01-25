package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type GroupBy struct {
	node

	Expressions []Expression
	Having      Expression
}

func (g *GroupBy) Accept(v AstVisitor) any {
	return v.VisitGroupBy(g)
}

func (g *GroupBy) Walk(w AstWalker) error {
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
