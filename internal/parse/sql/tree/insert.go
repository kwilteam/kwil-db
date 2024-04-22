package tree

import (
	"fmt"

	sqlwriter "github.com/kwilteam/kwil-db/internal/parse/sql/tree/sql-writer"
)

type InsertStmt struct {
	node

	CTE  []*CTE
	Core *InsertCore
}

func (ins *InsertStmt) Accept(v AstVisitor) any {
	return v.VisitInsertStmt(ins)
}

func (ins *InsertStmt) Walk(w AstListener) error {
	return run(
		w.EnterInsertStmt(ins),
		walkMany(w, ins.CTE),
		walk(w, ins.Core),
		w.ExitInsertStmt(ins),
	)
}

func (ins *InsertStmt) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	if len(ins.CTE) > 0 {
		stmt.Token.With()
		stmt.WriteList(len(ins.CTE), func(i int) {
			stmt.WriteString(ins.CTE[i].ToSQL())
		})
	}

	stmt.WriteString(ins.Core.ToSQL())

	stmt.Token.Semicolon()

	return stmt.String()
}

func (ins *InsertStmt) statement() {}

type InsertCore struct {
	node

	schema          string
	InsertType      InsertType
	Table           string
	TableAlias      string
	Columns         []string
	Values          [][]Expression
	Upsert          *Upsert
	ReturningClause *ReturningClause
}

func (ins *InsertCore) Accept(v AstVisitor) any {
	return v.VisitInsertCore(ins)
}

func (ins *InsertCore) Walk(w AstListener) error {
	return run(
		w.EnterInsertCore(ins),
		func() error {
			for _, v := range ins.Values {
				err := walkMany(w, v)
				if err != nil {
					return err
				}
			}
			return nil
		}(),
		walk(w, ins.Upsert),
		walk(w, ins.ReturningClause),
		w.ExitInsertCore(ins),
	)
}

// SetSchema sets the schema of the table.
// It should not be called by the parser, and is meant to be called
// by processes after parsing.
func (ins *InsertCore) SetSchema(schema string) {
	ins.schema = schema
}

type InsertType uint8

const (
	InsertTypeInsert InsertType = iota
)

func (i InsertType) Valid() error {
	switch i {
	case InsertTypeInsert:
		return nil
	default:
		return fmt.Errorf("invalid insert type: %d", i)
	}
}

func (i *InsertType) String() string {
	switch *i {
	case InsertTypeInsert:
		return "INSERT"
	default:
		panic(fmt.Errorf("unknown InsertType: %d", *i))
	}
}

func (ins *InsertCore) ToSQL() string {
	ins.check()

	stmt := sqlwriter.NewWriter()
	stmt.WriteString(ins.InsertType.String())
	stmt.Token.Into()

	if ins.schema != "" {
		stmt.Token.Space()
		stmt.WriteIdentNoSpace(ins.schema)
		stmt.Token.Period()
		stmt.WriteIdentNoSpace(ins.Table)
		stmt.Token.Space()
	} else {
		stmt.WriteIdent(ins.Table)
	}

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

func (ins *InsertCore) check() {
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
