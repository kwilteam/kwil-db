package tree

// UpdateSetClause is a clause that represents the SET clause in an UPDATE statement.
// This does NOT include the SET keyword.
// e.g. column1 = expression, column2 = expression, ...
type UpdateSetClause struct {
	stmt       *updateSetClauseBuilder
	Columns    []string
	Expression Expression
}

func (u *UpdateSetClause) ToSQL() string {
	u.stmt = Builder.BeginUpdateSetClause()

	if len(u.Columns) == 0 {
		panic("no column names provided to UpdateSetClause")
	} else if len(u.Columns) == 1 {
		u.stmt.ColumnName(u.Columns[0])
	} else {
		u.stmt.ColumnNameList(u.Columns)
	}

	if u.Expression == nil {
		panic("no expression provided to UpdateSetClause")
	}

	u.stmt.Expression(u.Expression)

	return u.stmt.String()
}

type updateSetClauseBuilder struct {
	stmt *sqlBuilder
}

func (b *builder) BeginUpdateSetClause() *updateSetClauseBuilder {
	u := &updateSetClauseBuilder{
		stmt: newSQLBuilder(),
	}

	return u
}

func (b *updateSetClauseBuilder) String() string {
	return b.stmt.String()
}

func (b *updateSetClauseBuilder) ColumnName(column string) {
	b.stmt.Write(SPACE)
	b.stmt.WriteIdent(column)
	b.stmt.Write(SPACE)
}

func (b *updateSetClauseBuilder) ColumnNameList(columns []string) {
	b.stmt.Write(SPACE, LPAREN)
	for i, col := range columns {
		if i > 0 && i < len(columns) {
			b.stmt.Write(COMMA, SPACE)
		}
		b.stmt.WriteIdent(col)
	}
	b.stmt.Write(RPAREN, SPACE)
}

func (b *updateSetClauseBuilder) Expression(expression Expression) {
	b.stmt.Write(SPACE, EQUALS, SPACE)
	b.stmt.WriteString(expression.ToSQL())
	b.stmt.Write(SPACE)
}
