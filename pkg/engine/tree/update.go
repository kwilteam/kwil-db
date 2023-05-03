package tree

type UpdateSetType uint8

const (
	UpdateSetTypeSetSingle UpdateSetType = iota // SET column1 = expression, column2 = expression, ...
	UpdateSetTypeSetList                        // SET column1, column2 = (SELECT ...)
)

// UpdateSetClause is a clause that represents the SET clause in an UPDATE statement.
// e.g. SET column1 = expression, column2 = expression, ...
type UpdateSetClause struct {
	Type UpdateSetType

	// SetSingle
	SetSingle *updateSetSingle

	// SetList
	SetList *updateSetList
}

type updateSetSingle struct {
	Sets map[string]Expression // map[columnName]expression
}

type updateSetList struct {
	Columns    []string
	Expression *Expression
}
