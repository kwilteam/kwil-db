package tree

type ReturningClause struct {
	stmt     *returningClauseBuilder
	Returned []*ReturningClauseColumn
}

func (r *ReturningClause) ToSQL() string {
	r.stmt = Builder.BeginReturningClause()
	r.stmt.Return(r.Returned)
	return r.stmt.String()
}

type ReturningClauseColumn struct {
	All        bool
	Expression Expression
	Alias      string
}

type returningClauseBuilder struct {
	stmt *sqlBuilder
}

func (b *builder) BeginReturningClause() *returningClauseBuilder {
	r := &returningClauseBuilder{
		stmt: newSQLBuilder(),
	}

	r.stmt.Write(SPACE, RETURNING, SPACE)

	return r
}

func (b *returningClauseBuilder) String() string {
	return b.stmt.String()
}

func (b *returningClauseBuilder) Return(rets []*ReturningClauseColumn) {
	b.stmt.Write(SPACE)
	for i, ret := range rets {
		if i > 0 && i < len(rets) {
			b.stmt.Write(COMMA, SPACE)
		}

		if ret.All {
			b.stmt.Write(ASTERISK)
		} else {
			b.stmt.WriteString(ret.Expression.ToSQL())
			if ret.Alias != "" {
				b.alias(ret.Alias)
			}
		}
	}

	b.stmt.Write(SPACE)
}

func (b *returningClauseBuilder) alias(alias string) {
	b.stmt.Write(SPACE, AS, SPACE)
	b.stmt.WriteIdent(alias)
	b.stmt.Write(SPACE)
}
