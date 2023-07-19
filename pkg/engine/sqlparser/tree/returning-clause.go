package tree

import (
	"errors"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"
)

type ReturningClause struct {
	Returned []*ReturningClauseColumn
}

func (r *ReturningClause) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitReturningClause(r),
		acceptMany(visitor, r.Returned),
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
	All        bool
	Expression Expression
	Alias      string
}

func (r *ReturningClauseColumn) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitReturningClauseColumn(r),
		accept(visitor, r.Expression),
	)
}
