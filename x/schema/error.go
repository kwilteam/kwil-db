package schema

import "errors"

var (
	ErrDatabaseNotFound = errors.New("database not found")
	ErrPlanNotFound     = errors.New("plan not found")
)
