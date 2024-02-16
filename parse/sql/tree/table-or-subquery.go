package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

// TableOrSubquery is any of:
//   - TableOrSubqueryTable
//   - TableOrSubquerySelect
//   - TableOrSubqueryList
type TableOrSubquery interface {
	ToSQL() string
	tableOrSubquery()
	Accept(w Walker) error
}

type TableOrSubqueryTable struct {
	schema string
	Name   string
	Alias  string
}

func (t *TableOrSubqueryTable) Accept(w Walker) error {
	return run(
		w.EnterTableOrSubqueryTable(t),
		w.ExitTableOrSubqueryTable(t),
	)
}

func (t *TableOrSubqueryTable) ToSQL() string {
	if t.Name == "" {
		panic("table name is empty")
	}

	stmt := sqlwriter.NewWriter()

	if t.schema != "" {
		stmt.Token.Space()
		stmt.WriteIdentNoSpace(t.schema)
		stmt.Token.Period()
		stmt.WriteIdentNoSpace(t.Name)
		stmt.Token.Space()
	} else {
		stmt.WriteIdentNoSpace(t.Name)
	}

	if t.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdentNoSpace(t.Alias)

	}

	return stmt.String()
}
func (t *TableOrSubqueryTable) tableOrSubquery() {}

// SetSchema sets the schema of the table.
// It should not be called by the parser, and is meant to be called
// by processes after parsing.
func (t *TableOrSubqueryTable) SetSchema(schema string) {
	t.schema = schema
}

type TableOrSubquerySelect struct {
	Select *SelectStmt
	Alias  string
}

func (t *TableOrSubquerySelect) Accept(w Walker) error {
	return run(
		w.EnterTableOrSubquerySelect(t),
		accept(w, t.Select),
		w.ExitTableOrSubquerySelect(t),
	)
}

func (t *TableOrSubquerySelect) ToSQL() string {
	if t.Select == nil {
		panic("select is nil")
	}

	stmt := sqlwriter.NewWriter()
	stmt.Token.Lparen()

	selectString := t.Select.ToSQL()
	stmt.WriteString(selectString)
	stmt.Token.Rparen()

	if t.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdent(t.Alias)

	}

	return stmt.String()
}
func (t *TableOrSubquerySelect) tableOrSubquery() {}

type TableOrSubqueryList struct {
	TableOrSubqueries []TableOrSubquery
}

func (t *TableOrSubqueryList) Accept(w Walker) error {
	return run(
		w.EnterTableOrSubqueryList(t),
		acceptMany(w, t.TableOrSubqueries),
		w.ExitTableOrSubqueryList(t),
	)
}

func (t *TableOrSubqueryList) ToSQL() string {
	if len(t.TableOrSubqueries) == 0 {
		panic("table or subquery list is empty")
	}

	stmt := sqlwriter.NewWriter()

	stmt.WriteParenList(len(t.TableOrSubqueries), func(i int) {
		stmt.WriteString(t.TableOrSubqueries[i].ToSQL())
	})

	return stmt.String()
}
func (t *TableOrSubqueryList) tableOrSubquery() {}

type TableOrSubqueryJoin struct {
	JoinClause *JoinClause
}

func (t *TableOrSubqueryJoin) Accept(w Walker) error {
	return run(
		w.EnterTableOrSubqueryJoin(t),
		accept(w, t.JoinClause),
		w.ExitTableOrSubqueryJoin(t),
	)
}

func (t *TableOrSubqueryJoin) tableOrSubquery() {}

func (t *TableOrSubqueryJoin) ToSQL() string {

	if t.JoinClause == nil {
		panic("join clause is nil")
	}

	stmt := sqlwriter.NewWriter()
	stmt.Token.Lparen()
	stmt.WriteString(t.JoinClause.ToSQL())
	stmt.Token.Rparen()

	return stmt.String()

}
