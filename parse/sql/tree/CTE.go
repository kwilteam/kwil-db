package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type CTE struct {
	*node

	Table   string
	Columns []string
	Select  *SelectStmt
}

func (c *CTE) Accept(v AstVisitor) any {
	return v.VisitCTE(c)
}

func (c *CTE) Walk(w AstWalker) error {
	return run(
		w.EnterCTE(c),
		accept(w, c.Select),
		w.ExitCTE(c),
	)
}

func (c *CTE) ToSQL() string {
	stmt := sqlwriter.NewWriter()
	stmt.WriteIdent(c.Table)

	if len(c.Columns) > 0 {
		stmt.WriteParenList(len(c.Columns), func(i int) {
			stmt.WriteIdent(c.Columns[i])
		})
	}

	stmt.Token.As()

	stmt.Token.Lparen()
	stmt.WriteString(c.Select.ToSQL())
	stmt.Token.Rparen()

	return stmt.String()
}
