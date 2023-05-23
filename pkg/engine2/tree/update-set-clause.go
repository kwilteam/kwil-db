package tree

import sqlwriter "github.com/kwilteam/kwil-db/pkg/engine2/tree/sql-writer"

// UpdateSetClause is a clause that represents the SET clause in an UPDATE statement.
// This does NOT include the SET keyword.
// e.g. column1 = expression, column2 = expression, ...
type UpdateSetClause struct {
	Columns    []string
	Expression Expression
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
