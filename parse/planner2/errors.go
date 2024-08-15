package planner2

import (
	"errors"
)

var (
	ErrIllegalAggregate = errors.New("illegal aggregate")
	ErrColumnNotFound   = errors.New("column not found or cannot be referenced in this part of the query")
)