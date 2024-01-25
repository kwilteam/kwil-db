package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// UpdateSetClause is a clause that represents the SET clause in an UPDATE statement.
// This does NOT include the SET keyword.
// e.g. column1 = expression, column2 = expression, ...
type UpdateSetClause struct {
	node

	Columns    []string
	Expression Expression
}

func (u *UpdateSetClause) Accept(v AstVisitor) any {
	return v.VisitUpdateSetClause(u)
}

func (u *UpdateSetClause) Walk(w AstWalker) error {
	return run(
		w.EnterUpdateSetClause(u),
		accept(w, u.Expression),
		w.ExitUpdateSetClause(u),
	)
}

func (u *UpdateSetClause) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(u.Columns) == 0 {
		panic("no column names provided to UpdateSetClause")
	} else if len(u.Columns) == 1 {
		stmt.WriteIdent(u.Columns[0])
	} else {
		stmt.WriteParenList(len(u.Columns), func(i int) {
			stmt.WriteIdent(u.Columns[i])
		})
	}

	if u.Expression == nil {
		panic("no expression provided to UpdateSetClause")
	}

	stmt.Token.Equals()
	stmt.WriteString(u.Expression.ToSQL())

	return stmt.String()
}
