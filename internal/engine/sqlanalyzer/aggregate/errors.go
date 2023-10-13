package aggregate

import "errors"

var (
	ErrAggregateFuncContainsSubquery       = errors.New("aggregate functions cannot contain subqueries")
	ErrAggregateFuncHasInvalidPosArg       = errors.New("aggregate function %s cannot contain a column in an argument that is not its first")
	ErrAggregateQueryContainsSelectAll     = errors.New("aggregate functions cannot be used with SELECT *, or any variant of SELECT *")
	ErrResultSetContainsBareColumn         = errors.New("aggregate queries must not return bare columns that are not encapsulated in an aggregate function, or included in the GROUP BY clause")
	ErrHavingClauseContainsUngroupedColumn = errors.New("having clause column %s must be in the GROUP BY clause")
	ErrGroupByContainsInvalidExpr          = errors.New("group by clause can only contain bare columns")
)
