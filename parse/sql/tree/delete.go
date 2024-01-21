package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type Delete struct {
	node

	CTE        []*CTE
	DeleteStmt *DeleteStmt
}

func (d *Delete) Accept(v AstVisitor) any {
	return v.VisitDelete(d)
}

func (d *Delete) Walk(w AstListener) error {
	return run(
		w.EnterDelete(d),
		walkMany(w, d.CTE),
		walk(w, d.DeleteStmt),
		w.ExitDelete(d),
	)
}

func (d *Delete) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(d.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(d.CTE), func(i int) {
			stmt.WriteString(d.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(d.DeleteStmt.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String()
}

type DeleteStmt struct {
	node

	QualifiedTableName *QualifiedTableName
	Where              Expression
	Returning          *ReturningClause
}

func (d *DeleteStmt) Accept(v AstVisitor) any {
	return v.VisitDeleteStmt(d)
}

func (d *DeleteStmt) Walk(w AstListener) error {
	return run(
		w.EnterDeleteStmt(d),
		walk(w, d.QualifiedTableName),
		walk(w, d.Where),
		walk(w, d.Returning),
		w.ExitDeleteStmt(d),
	)
}

func (d *DeleteStmt) ToSQL() string {
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

func (d *DeleteStmt) check() {
	if d.QualifiedTableName == nil {
		panic("qualified table name is nil")
	}
}
