package tree

type sqlable interface {
	ToSQL() (string, error)
}

// Statement is a generalized wrapper for any full statement (e.g. SELECT, INSERT, UPDATE, DELETE) that can use a common table expression
type Statement[T sqlable] struct {
	CTE       []*CTE
	Statement T
}
