package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type ReturningClause struct {
	node

	Returned []*ReturningClauseColumn
}

func (r *ReturningClause) Accept(v AstVisitor) any {
	return v.VisitReturningClause(r)
}

func (r *ReturningClause) Walk(w AstWalker) error {
	return run(
		w.EnterReturningClause(r),
		acceptMany(w, r.Returned),
		w.ExitReturningClause(r),
	)
}

func (r *ReturningClause) ToSQL() string {
	r.check()

	stmt := sqlwriter.NewWriter()
	stmt.Token.Returning()

	stmt.WriteList(len(r.Returned), func(i int) {
		if r.Returned[i].All {
			stmt.Token.Asterisk()
		} else {
			stmt.WriteString(r.Returned[i].Expression.ToSQL())
			if r.Returned[i].Alias != "" {
				stmt.Token.As()
				stmt.WriteIdent(r.Returned[i].Alias)
			}
		}
	})

	return stmt.String()
}

func (r *ReturningClause) check() {
	if len(r.Returned) == 0 {
		panic("no columns provided to ReturningClause")
	}

	for _, col := range r.Returned {
		if col.All && col.Expression != nil {
			panic("all and expression cannot be set at the same time")
		}
	}
}

type ReturningClauseColumn struct {
	node

	All        bool
	Expression Expression
	Alias      string
}

func (r *ReturningClauseColumn) Accept(v AstVisitor) any {
	return v.VisitReturningClauseColumn(r)
}

func (r *ReturningClauseColumn) Walk(w AstWalker) error {
	return run(
		w.EnterReturningClauseColumn(r),
		accept(w, r.Expression),
		w.ExitReturningClauseColumn(r),
	)
}
