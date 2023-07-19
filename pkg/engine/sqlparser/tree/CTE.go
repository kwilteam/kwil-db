package tree

import (
	"errors"

	sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree/sql-writer"
)

type CTE struct {
	Table   string
	Columns []string
	Select  *SelectStmt
}

func (c *CTE) Accept(visitor Visitor) error {
	return errors.Join(
		visitor.VisitCTE(c),
		accept(visitor, c.Select),
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
