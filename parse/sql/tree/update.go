package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// Update Statement with CTEs
type Update struct {
	CTE        []*CTE
	UpdateStmt *UpdateStmt
}

func (u *Update) Accept(w Walker) error {
	return run(
		w.EnterUpdate(u),
		acceptMany(w, u.CTE),
		accept(w, u.UpdateStmt),
		w.ExitUpdate(u),
	)
}

func (u *Update) ToSQL() (str string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err2, ok := r.(error)
			if !ok {
				err2 = fmt.Errorf("%v", r)
			}

			err = err2
		}
	}()

	stmt := sqlwriter.NewWriter()

	if len(u.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(u.CTE), func(i int) {
			stmt.WriteString(u.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(u.UpdateStmt.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String(), nil
}

// UpdateStmt is a statement that represents an UPDATE statement.
// USE Update INSTEAD OF THIS
type UpdateStmt struct {
	QualifiedTableName *QualifiedTableName
	UpdateSetClause    []*UpdateSetClause
	From               *FromClause
	Where              Expression
	Returning          *ReturningClause
}

func (u *UpdateStmt) Accept(w Walker) error {
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

func (u *UpdateStmt) ToSQL() string {
	u.check()

	stmt := sqlwriter.NewWriter()
	stmt.Token.Update()
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
