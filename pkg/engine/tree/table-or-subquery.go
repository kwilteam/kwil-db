package tree

type TableOrSubquery interface {
	ToSQL() string
	TableOrSubquery()
}

type TableOrSubqueryTable struct {
	Name  string
	Alias string
}

func (t *TableOrSubqueryTable) ToSQL() string {
	if t.Name == "" {
		panic("table name is empty")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE)
	stmt.WriteString(t.Name)
	stmt.Write(SPACE)

	if t.Alias != "" {
		stmt.Write(SPACE, AS, SPACE)
		stmt.WriteString(t.Alias)
		stmt.Write(SPACE)
	}

	return stmt.String()
}
func (t *TableOrSubqueryTable) TableOrSubquery() {}

type TableOrSubquerySelect struct {
	Select *Select
	Alias  string
}

func (t *TableOrSubquerySelect) ToSQL() string {
	if t.Select == nil {
		panic("select is nil")
	}

	stmt := newSQLBuilder()
	stmt.Write(SPACE, LPAREN, SPACE)

	selectString, err := t.Select.ToSQL()
	if err != nil {
		panic(err)
	}
	stmt.WriteString(selectString)
	stmt.Write(SPACE, RPAREN, SPACE)

	if t.Alias != "" {
		stmt.Write(SPACE, AS, SPACE)
		stmt.WriteString(t.Alias)
		stmt.Write(SPACE)
	}

	return stmt.String()
}
func (t *TableOrSubquerySelect) TableOrSubquery() {}

type TableOrSubqueryList struct {
	TableOrSubqueries []TableOrSubquery
}

func (t *TableOrSubqueryList) ToSQL() string {
	if len(t.TableOrSubqueries) == 0 {
		panic("table or subquery list is empty")
	}

	stmt := newSQLBuilder()

	stmt.Write(SPACE, LPAREN, SPACE)
	for i, tableOrSubquery := range t.TableOrSubqueries {
		if i > 0 && i < len(t.TableOrSubqueries) {
			stmt.Write(COMMA, SPACE)
		}

		stmt.WriteString(tableOrSubquery.ToSQL())
	}
	stmt.Write(SPACE, RPAREN, SPACE)

	return stmt.String()
}
func (t *TableOrSubqueryList) TableOrSubquery() {}

type TableOrSubqueryJoin struct {
	JoinClause *JoinClause
}

func (t *TableOrSubqueryJoin) ToSQL() string {

	/*
		if t.JoinClause == nil {
			panic("join clause is nil")
		}

		stmt := newSQLBuilder()
		stmt.Write(SPACE, LPAREN, SPACE)
		stmt.WriteString(t.JoinClause.ToSQL())
		stmt.Write(SPACE, RPAREN, SPACE)

		return stmt.String()
	*/
	return ""
}
