package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type ConflictTarget struct {
	node

	IndexedColumns []string
	Where          Expression
}

func (c *ConflictTarget) Accept(v AstVisitor) any {
	return v.VisitConflictTarget(c)
}

func (c *ConflictTarget) Walk(w AstListener) error {
	return run(
		w.EnterConflictTarget(c),
		walk(w, c.Where),
		w.ExitConflictTarget(c),
	)
}

func (c *ConflictTarget) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(c.IndexedColumns) > 0 {
		stmt.WriteParenList(len(c.IndexedColumns), func(i int) {
			stmt.WriteIdent(c.IndexedColumns[i])
		})
	}

	if c.Where != nil {
		stmt.Token.Where()
		stmt.WriteString(c.Where.ToSQL())
	}

	return stmt.String()
}
