package tree

import sqlwriter "github.com/kwilteam/kwil-db/pkg/engine2/tree/sql-writer"

// Update Statement with CTEs
type Update struct {
	CTE        []*CTE
	UpdateStmt *UpdateStmt
}

func (u *Update) ToSQL() (str string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
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
	Or                 UpdateOr
	QualifiedTableName *QualifiedTableName
	UpdateSetClause    []*UpdateSetClause
	From               *FromClause
	Where              Expression
	Returning          *ReturningClause
}

type UpdateOr string

const (
	UpdateOrAbort    UpdateOr = "ABORT"
	UpdateOrFail     UpdateOr = "FAIL"
	UpdateOrIgnore   UpdateOr = "IGNORE"
	UpdateOrReplace  UpdateOr = "REPLACE"
	UpdateOrRollback UpdateOr = "ROLLBACK"
)

func (u *UpdateOr) check() {
	switch *u {
	case UpdateOrAbort:
	case UpdateOrFail:
	case UpdateOrIgnore:
	case UpdateOrReplace:
	case UpdateOrRollback:
	default:
		panic("unknown UpdateOr")
	}
}

func (u *UpdateOr) ToSQL() string {
	u.check()

	stmt := sqlwriter.NewWriter()
	stmt.Token.Or()
	stmt.WriteString(string(*u))
	return stmt.String()
}

func (u *UpdateStmt) ToSQL() string {
	u.check()

	stmt := sqlwriter.NewWriter()
	stmt.Token.Update()
	if u.Or != "" {
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
