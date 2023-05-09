package tree

type OrderBy struct {
	OrderingTerms []*OrderingTerm
}

func (o *OrderBy) ToSQL() string {
	stmt := newSQLBuilder()

	stmt.Write(SPACE, ORDER, SPACE, BY, SPACE)

	if len(o.OrderingTerms) == 0 {
		panic("no ordering terms provided to OrderBy")
	}

	for i, term := range o.OrderingTerms {
		if i > 0 && i < len(o.OrderingTerms) {
			stmt.Write(COMMA, SPACE)
		}

		stmt.WriteString(term.ToSQL())
	}

	return stmt.String()
}

type OrderingTerm struct {
	Expression   Expression
	Collation    CollationType
	OrderType    OrderType
	NullOrdering NullOrderingType
}

func (o *OrderingTerm) ToSQL() string {
	stmt := newSQLBuilder()

	stmt.WriteString(o.Expression.ToSQL())

	if o.Collation != CollationTypeNone {
		stmt.Write(SPACE, COLLATE, SPACE)
		stmt.WriteString(o.Collation.String())
		stmt.Write(SPACE)
	}

	if o.OrderType != OrderTypeNone {
		stmt.Write(SPACE)
		stmt.WriteString(o.OrderType.String())
		stmt.Write(SPACE)
	}

	if o.NullOrdering != NullOrderingTypeNone {
		stmt.Write(SPACE)
		stmt.WriteString(o.NullOrdering.String())
		stmt.Write(SPACE)
	}

	return stmt.String()
}

type NullOrderingType string

const (
	NullOrderingTypeNone  NullOrderingType = ""
	NullOrderingTypeFirst NullOrderingType = "NULLS FIRST"
	NullOrderingTypeLast  NullOrderingType = "NULLS LAST"
)

func (n NullOrderingType) String() string {
	return string(n)
}
