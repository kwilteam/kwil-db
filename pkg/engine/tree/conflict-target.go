package tree

type ConflictTarget struct {
	stmt           *conflictTargetBuilder
	IndexedColumns []*IndexedColumn
	Where          Expression
}

func (c *ConflictTarget) ToSQL() string {
	c.stmt = Builder.BeginConflictTarget()

	if len(c.IndexedColumns) > 0 {
		c.stmt.IndexedColumns(c.IndexedColumns)
	}

	if c.Where != nil {
		c.stmt.Where(c.Where)
	}

	return c.stmt.String()
}

type conflictTargetBuilder struct {
	stmt *sqlBuilder
}

func (b *builder) BeginConflictTarget() *conflictTargetBuilder {
	return &conflictTargetBuilder{
		stmt: newSQLBuilder(),
	}
}

func (b *conflictTargetBuilder) IndexedColumns(ics []*IndexedColumn) {
	b.stmt.Write(SPACE, LPAREN)
	for i, ic := range ics {
		if i > 0 && i < len(ics) {
			b.stmt.Write(COMMA, SPACE)
		}
		b.stmt.WriteString(ic.ToSQL())
	}
	b.stmt.Write(RPAREN, SPACE)
}

func (b *conflictTargetBuilder) Where(w Expression) {
	b.stmt.Write(SPACE, WHERE, SPACE)
	b.stmt.WriteString(w.ToSQL())
}

func (b *conflictTargetBuilder) String() string {
	return b.stmt.String()
}
