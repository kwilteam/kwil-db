package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// Update Statement with CTEs
type Update struct {
	node

	CTE        []*CTE
	UpdateStmt *UpdateStmt
}

func (u *Update) Accept(v AstVisitor) any {
	return v.VisitUpdate(u)
}

func (u *Update) Walk(w AstWalker) error {
	return run(
		w.EnterUpdate(u),
		acceptMany(w, u.CTE),
		accept(w, u.UpdateStmt),
		w.ExitUpdate(u),
	)
}

func (u *Update) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(u.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(u.CTE), func(i int) {
			stmt.WriteString(u.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(u.UpdateStmt.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String()
}

// UpdateStmt is a statement that represents an UPDATE statement.
// USE Update INSTEAD OF THIS
type UpdateStmt struct {
	node

	Or                 UpdateOr
	QualifiedTableName *QualifiedTableName
	UpdateSetClause    []*UpdateSetClause
	From               *FromClause
	Where              Expression
	Returning          *ReturningClause
}

func (u *UpdateStmt) Accept(v AstVisitor) any {
	return v.VisitUpdateStmt(u)
}

func (u *UpdateStmt) Walk(w AstWalker) error {
	return run(
		w.EnterUpdateStmt(u),
		accept(w, u.QualifiedTableName),
		acceptMany(w, u.UpdateSetClause),
		accept(w, u.From),
		accept(w, u.Where),
		accept(w, u.Returning),
		w.ExitUpdateStmt(u),
	)
}

type UpdateOr string

const (
	UpdateOrAbort    UpdateOr = "ABORT"
	UpdateOrFail     UpdateOr = "FAIL"
	UpdateOrIgnore   UpdateOr = "IGNORE"
	UpdateOrReplace  UpdateOr = "REPLACE"
	UpdateOrRollback UpdateOr = "ROLLBACK"
)

func (u *UpdateOr) Valid() error {
	if u.Empty() {
		return nil
	}

	switch *u {
	case UpdateOrAbort, UpdateOrFail, UpdateOrIgnore, UpdateOrReplace, UpdateOrRollback:
		return nil
	default:
		return fmt.Errorf("unknown UpdateOr: %s", *u)
	}
}

func (u UpdateOr) Empty() bool {
	return u == ""
}

func (u UpdateOr) check() {
	if u.Empty() {
		return
	}
	if err := u.Valid(); err != nil {
		panic(err)
	}
}

func (u *UpdateOr) ToSQL() string {
	u.check()
	if u.Empty() {
		return ""
	}

	stmt := sqlwriter.NewWriter()
	stmt.Token.Or()
	stmt.WriteString(string(*u))
	return stmt.String()
}

func (u *UpdateStmt) ToSQL() string {
	u.check()

	stmt := sqlwriter.NewWriter()
	stmt.Token.Update()
	if !u.Or.Empty() {
		stmt.WriteString(u.Or.ToSQL())
	}
	stmt.WriteString(u.QualifiedTableName.ToSQL())
	stmt.Token.Set()
	stmt.WriteList(len(u.UpdateSetClause), func(i int) {
		stmt.WriteString(u.UpdateSetClause[i].ToSQL())
	})

	if u.From != nil {
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

func (u *UpdateStmt) check() {
	if u.QualifiedTableName == nil {
		panic("qualified table name is required")
	}
	if len(u.UpdateSetClause) == 0 {
		panic("update set clause is required")
	}
}
