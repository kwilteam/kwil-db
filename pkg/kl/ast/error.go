package ast

import "errors"

var (
	ErrMultiplePrimaryKeys        = errors.New("multiple primary keys")
	ErrDuplicateColumnOrIndexName = errors.New("duplicate column or index")
	ErrDuplicateTableName         = errors.New("duplicate table")
	ErrDuplicateActionName        = errors.New("duplicate action")
	ErrInvalidActionParam         = errors.New("invalid action param")
	ErrInvalidColumnName          = errors.New("invalid column name")
	ErrInvalidStatement           = errors.New("invalid statement")
)
