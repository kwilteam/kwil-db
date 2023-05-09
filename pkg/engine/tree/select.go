package tree

type Select struct {
	SelectClause *SelectClause
	OrderBy      *OrderBy
	Limit        *Limit
}

func (s *Select) ToSQL() (string, error) {
	return "", nil
}

type SelectClause struct {
	SelectType SelectType
	Columns    []string
	From       *FromClause
	Where      *Expression
	GroupBy    *GroupBy
	Compound   *CompoundOperator
}

func (s *SelectClause) ToSQL() string {
	return ""
}

type SelectType uint8

const (
	SelectTypeAll SelectType = iota
	SelectTypeDistinct
)

type FromClause struct {
	TableOrSubquery TableOrSubquery
	JoinClauses     []*JoinClause
}

type CompoundOperatorType uint8

const (
	CompoundOperatorTypeUnion CompoundOperatorType = iota
	CompoundOperatorTypeUnionAll
	CompoundOperatorTypeIntersect
	CompoundOperatorTypeExcept
)

type CompoundOperator struct {
	Operator     CompoundOperatorType
	SelectClause *SelectClause
}
