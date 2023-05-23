package tree

import sqlwriter "github.com/kwilteam/kwil-db/pkg/engine2/tree/sql-writer"

// TableOrSubquery is any of:
//   - TableOrSubqueryTable
//   - TableOrSubquerySelect
//   - TableOrSubqueryList
type TableOrSubquery interface {
	ToSQL() string
	tableOrSubquery()
}

type TableOrSubqueryTable struct {
	Name  string
	Alias string
}

func (t *TableOrSubqueryTable) ToSQL() string {
	if t.Name == "" {
		panic("table name is empty")
	}

	stmt := sqlwriter.NewWriter()
	stmt.WriteIdentNoSpace(t.Name)

	if t.Alias != "" {
		stmt.Token.As()
		stmt.WriteIdentNoSpace(t.Alias)

	}

	return stmt.String()
}
func (t *TableOrSubqueryTable) tableOrSubquery() {}

type TableOrSubquerySelect struct {
	Select *SelectStmt
	Alias  string
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
		stmt.WriteString(t.Alias)

	}

	return stmt.String()
}
func (t *TableOrSubquerySelect) tableOrSubquery() {}

type TableOrSubqueryList struct {
	TableOrSubqueries []TableOrSubquery
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
