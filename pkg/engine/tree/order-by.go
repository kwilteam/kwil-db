package tree

import sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/tree/sql-writer"

type OrderBy struct {
	OrderingTerms []*OrderingTerm
}

func (o *OrderBy) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	stmt.Token.Order().By()

	if len(o.OrderingTerms) == 0 {
		panic("no ordering terms provided to OrderBy")
	}

	stmt.WriteList(len(o.OrderingTerms), func(i int) {
		stmt.WriteString(o.OrderingTerms[i].ToSQL())
	})

	return stmt.String()
}

type OrderingTerm struct {
	Expression   Expression
	Collation    CollationType
	OrderType    OrderType
	NullOrdering NullOrderingType
}

func (o *OrderingTerm) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	stmt.WriteString(o.Expression.ToSQL())

	if o.Collation.Valid() {
		stmt.Token.Collate()
		stmt.WriteString(o.Collation.String())
	}

	if o.OrderType != OrderTypeNone {
		stmt.WriteString(o.OrderType.String())
	}

	if o.NullOrdering != NullOrderingTypeNone {
		stmt.WriteString(o.NullOrdering.String())
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
