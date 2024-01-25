package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type Insert struct {
	node

	CTE        []*CTE
	InsertStmt *InsertStmt
}

func (ins *Insert) Accept(v AstVisitor) any {
	return v.VisitInsert(ins)
}

func (ins *Insert) Walk(w AstWalker) error {
	return run(
		w.EnterInsert(ins),
		acceptMany(w, ins.CTE),
		accept(w, ins.InsertStmt),
		w.ExitInsert(ins),
	)
}

func (ins *Insert) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(ins.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(ins.CTE), func(i int) {
			stmt.WriteString(ins.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(ins.InsertStmt.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String()
}

type InsertStmt struct {
	node

	InsertType      InsertType
	Table           string
	TableAlias      string
	Columns         []string
	Values          [][]Expression
	Upsert          *Upsert
	ReturningClause *ReturningClause
}

func (ins *InsertStmt) Accept(v AstVisitor) any {
	return v.VisitInsertStmt(ins)
}

func (ins *InsertStmt) Walk(w AstWalker) error {
	return run(
		w.EnterInsertStmt(ins),
		func() error {
			for _, v := range ins.Values {
				err := acceptMany(w, v)
				if err != nil {
					return err
				}
			}
			return nil
		}(),
		accept(w, ins.Upsert),
		accept(w, ins.ReturningClause),
		w.ExitInsertStmt(ins),
	)
}

type InsertType uint8

const (
	InsertTypeInsert InsertType = iota
	InsertTypeReplace
	InsertTypeInsertOrReplace
)

func (i InsertType) Valid() error {
	switch i {
	case InsertTypeInsert, InsertTypeReplace, InsertTypeInsertOrReplace:
		return nil
	default:
		return fmt.Errorf("invalid insert type: %d", i)
	}
}

func (i *InsertType) String() string {
	switch *i {
	case InsertTypeInsert:
		return "INSERT"
	case InsertTypeReplace:
		return "REPLACE"
	case InsertTypeInsertOrReplace:
		return "INSERT OR REPLACE"
	default:
		panic(fmt.Errorf("unknown InsertType: %d", *i))
	}
}

func (ins *InsertStmt) ToSQL() string {
	ins.check()

	stmt := sqlwriter.NewWriter()
	stmt.WriteString(ins.InsertType.String())
	stmt.Token.Into()
	stmt.WriteIdent(ins.Table)

	if ins.TableAlias != "" {
		stmt.Token.As()
		stmt.WriteIdent(ins.TableAlias)
	}
	if len(ins.Columns) > 0 {
		stmt.WriteParenList(len(ins.Columns), func(i int) {
			stmt.WriteIdent(ins.Columns[i])
		})
	}

	stmt.Token.Values()
	for i := range ins.Values {
		if i > 0 && i < len(ins.Values) {
			stmt.Token.Comma()
		}

		stmt.WriteParenList(len(ins.Values[i]), func(j int) {
			stmt.WriteString(ins.Values[i][j].ToSQL())
		})
	}

	if ins.Upsert != nil {
		stmt.WriteString(ins.Upsert.ToSQL())
	}

	if ins.ReturningClause != nil {
		stmt.WriteString(ins.ReturningClause.ToSQL())
	}

	return stmt.String()
}

func (ins *InsertStmt) check() {
	if ins.Table == "" {
		panic("InsertStatement: table name is empty")
	}

	if len(ins.Values) == 0 {
		panic("InsertStatement: values is empty")
	}

	if ins.Upsert != nil && ins.InsertType != InsertTypeInsert {
		panic("InsertStatement: upsert is only allowed for INSERT")
	}
}
