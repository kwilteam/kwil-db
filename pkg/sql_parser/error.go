package sql_parser

import (
	"errors"
)

var (
	ErrSyntax = errors.New("syntax error")

	ErrTableNotFound  = errors.New("table not found")
	ErrColumnNotFound = errors.New("column not found")
)
