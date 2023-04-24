package tree

type SelectStatement struct {
	CTEs       []*CTE
	SelectType SelectType
	Columns    []string
	From       *FromClause
}

type SelectType uint8

const (
	SelectTypeAll SelectType = iota
	SelectTypeDistinct
)

type FromClause struct {
	TableOrSubquery *TableOrSubquery
	JoinClauses     []*JoinClause
}

type WhereClause struct{}
