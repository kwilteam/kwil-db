package logical

import (
	"errors"
)

var (
	ErrIllegalAggregate           = errors.New("illegal aggregate")
	ErrIllegalWindowFunction      = errors.New("illegal window function")
	ErrColumnNotFound             = errors.New("column not found or cannot be referenced in this part of the query")
	ErrUpdateOrDeleteWithoutWhere = errors.New("UPDATE and DELETE statements with a FROM table require a WHERE clause")
	ErrUnknownTable               = errors.New("unknown table")
	ErrSetIncompatibleSchemas     = errors.New("incompatible schemas in COMPOUND operation")
	ErrNotNullableColumn          = errors.New("column is not nullable")
	ErrIllegalConflictArbiter     = errors.New("illegal conflict arbiter")
	ErrAmbiguousColumn            = errors.New("ambiguous column")
	ErrWindowAlreadyDefined       = errors.New("window already defined")
	ErrInvalidWindowFunction      = errors.New("invalid window function")
	ErrWindowNotDefined           = errors.New("window not defined")
	ErrFunctionDoesNotExist       = errors.New("function does not exist")
)
