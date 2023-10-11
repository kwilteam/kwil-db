package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type Delete struct {
	CTE        []*CTE
	DeleteStmt *DeleteStmt
}

func (d *Delete) Accept(w Walker) error {
	return run(
		w.EnterDelete(d),
		acceptMany(w, d.CTE),
		accept(w, d.DeleteStmt),
		w.ExitDelete(d),
	)
}

func (d *Delete) ToSQL() (str string, err error) {
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

	if len(d.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(d.CTE), func(i int) {
			stmt.WriteString(d.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(d.DeleteStmt.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String(), nil
}

type DeleteStmt struct {
	QualifiedTableName *QualifiedTableName
	Where              Expression
	Returning          *ReturningClause
}

func (d *DeleteStmt) Accept(w Walker) error {
	return run(
		w.EnterDeleteStmt(d),
		accept(w, d.QualifiedTableName),
		accept(w, d.Where),
		accept(w, d.Returning),
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
