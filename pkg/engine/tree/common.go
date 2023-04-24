package tree

type Alias struct{}

type ReturningClause struct{}

type CTE struct {
	Table   string
	Columns []string
	Select  *SelectStatement
}

type TableOrSubquery struct{}

type JoinClause struct{}
