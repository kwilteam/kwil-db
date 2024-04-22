package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/internal/parse/sql/tree/sql-writer"
)

type UpdateStmt struct {
	node

	CTE  []*CTE
	Core *UpdateCore
}

func (u *UpdateStmt) Accept(v AstVisitor) any {
	return v.VisitUpdateStmt(u)
}

func (u *UpdateStmt) Walk(w AstListener) error {
	return run(
		w.EnterUpdateStmt(u),
		walkMany(w, u.CTE),
		walk(w, u.Core),
		w.ExitUpdateStmt(u),
	)
}

func (u *UpdateStmt) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(u.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(u.CTE), func(i int) {
			stmt.WriteString(u.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(u.Core.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String()
}

func (u *UpdateStmt) statement() {}

type UpdateCore struct {
	node

	QualifiedTableName *QualifiedTableName
	UpdateSetClause    []*UpdateSetClause
	From               Relation
	Where              Expression
	Returning          *ReturningClause
}

func (u *UpdateCore) Accept(v AstVisitor) any {
	return v.VisitUpdateCore(u)
}

func (u *UpdateCore) Walk(w AstListener) error {
	return run(
		w.EnterUpdateCore(u),
		walk(w, u.QualifiedTableName),
		walkMany(w, u.UpdateSetClause),
		walk(w, u.From),
		walk(w, u.Where),
		walk(w, u.Returning),
		w.ExitUpdateCore(u),
	)
}

func (u *UpdateCore) ToSQL() string {
	u.check()

	stmt := sqlwriter.NewWriter()
	stmt.Token.Update()
	stmt.WriteString(u.QualifiedTableName.ToSQL())
	stmt.Token.Set()
	stmt.WriteList(len(u.UpdateSetClause), func(i int) {
		stmt.WriteString(u.UpdateSetClause[i].ToSQL())
	})

	if u.From != nil {
		stmt.Token.From()
		stmt.WriteString(u.From.ToSQL())
	}

	if u.Where != nil {
		stmt.Token.Where()
		stmt.WriteString(u.Where.ToSQL())
	}

	if u.Returning != nil {
		stmt.WriteString(u.Returning.ToSQL())
	}

	return stmt.String()
}

func (u *UpdateCore) check() {
	if u.QualifiedTableName == nil {
		panic("qualified table name is required")
	}
	if len(u.UpdateSetClause) == 0 {
		panic("update set clause is required")
	}
}
