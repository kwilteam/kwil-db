package planner

import (
	"errors"
)

var (
	ErrIllegalAggregate           = errors.New("illegal aggregate")
	ErrColumnNotFound             = errors.New("column not found or cannot be referenced in this part of the query")
	ErrUpdateOrDeleteWithoutWhere = errors.New("UPDATE and DELETE statements with a FROM table require a WHERE clause")
	ErrUnknownTable               = errors.New("unknown table")
	ErrAggregateInWhere           = errors.New("aggregate functions are not allowed in the WHERE clause")
	ErrSetIncompatibleSchemas     = errors.New("incompatible schemas in COMPOUND operation")
	ErrNotNullableColumn          = errors.New("column is not nullable")
	ErrIllegalConflictArbiter     = errors.New("illegal conflict arbiter")
	ErrAmbiguousColumn            = errors.New("ambiguous column")
)
