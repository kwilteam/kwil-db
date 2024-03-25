package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type DeleteStmt struct {
	node

	CTE  []*CTE
	Core *DeleteCore
}

func (d *DeleteStmt) Accept(v AstVisitor) any {
	return v.VisitDeleteStmt(d)
}

func (d *DeleteStmt) Walk(w AstListener) error {
	return run(
		w.EnterDeleteStmt(d),
		walkMany(w, d.CTE),
		walk(w, d.Core),
		w.ExitDeleteStmt(d),
	)
}

func (d *DeleteStmt) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(d.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(d.CTE), func(i int) {
			stmt.WriteString(d.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(d.Core.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String()
}

func (d *DeleteStmt) statement() {}

type DeleteCore struct {
	node

	QualifiedTableName *QualifiedTableName
	Where              Expression
	Returning          *ReturningClause
}

func (d *DeleteCore) Accept(v AstVisitor) any {
	return v.VisitDeleteCore(d)
}

func (d *DeleteCore) Walk(w AstListener) error {
	return run(
		w.EnterDeleteCore(d),
		walk(w, d.QualifiedTableName),
		walk(w, d.Where),
		walk(w, d.Returning),
		w.ExitDeleteCore(d),
	)
}

func (d *DeleteCore) ToSQL() string {
	d.check()

	stmt := sqlwriter.NewWriter()
	stmt.Token.Delete().From()
	stmt.WriteString(d.QualifiedTableName.ToSQL())
	if d.Where != nil {
		stmt.Token.Where()
		stmt.WriteString(d.Where.ToSQL())
	}
	if d.Returning != nil {
		stmt.WriteString(d.Returning.ToSQL())
	}

	return stmt.String()
}

func (d *DeleteCore) check() {
	if d.QualifiedTableName == nil {
		panic("qualified table name is nil")
	}
}
