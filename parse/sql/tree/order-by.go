package tree

import (
	"fmt"
	"strings"

	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type OrderBy struct {
	node

	OrderingTerms []*OrderingTerm
}

func (o *OrderBy) Accept(v AstVisitor) any {
	return v.VisitOrderBy(o)
}

func (o *OrderBy) Walk(w AstListener) error {
	return run(
		w.EnterOrderBy(o),
		walkMany(w, o.OrderingTerms),
		w.ExitOrderBy(o),
	)
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
	node

	Expression   Expression
	OrderType    OrderType
	NullOrdering NullOrderingType
}

func (o *OrderingTerm) Accept(v AstVisitor) any {
	return v.VisitOrderingTerm(o)
}

func (o *OrderingTerm) Walk(w AstListener) error {
	return run(
		w.EnterOrderingTerm(o),
		walk(w, o.Expression),
		w.ExitOrderingTerm(o),
	)
}

func (o *OrderingTerm) ToSQL() string {
	stmt := sqlwriter.NewWriter()

	stmt.WriteString(o.Expression.ToSQL())

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

func (n NullOrderingType) Valid() error {
	if n != NullOrderingTypeFirst && n != NullOrderingTypeLast && n != NullOrderingTypeNone {
		return fmt.Errorf("invalid null ordering type: %s", n)
	}

	return nil
}

type OrderType string

const (
	OrderTypeNone OrderType = ""
	OrderTypeAsc  OrderType = "ASC"
	OrderTypeDesc OrderType = "DESC"
)

func (o OrderType) String() string {
	o.check()
	return string(o)
}

func (o OrderType) check() {

	err := o.Valid()
	if err != nil {
		panic(err)
	}
}

func (o *OrderType) Valid() error {
	upper := OrderType(strings.ToUpper(string(*o)))

	if upper != OrderTypeAsc && upper != OrderTypeDesc && upper != OrderTypeNone {
		return fmt.Errorf("invalid order type: %s", o)
	}

	*o = upper

	return nil
}
