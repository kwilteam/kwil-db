package clean

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidIdentifier               = errors.New("invalid identifier")
	ErrInvalidLiteral                  = errors.New("invalid literal")
	ErrInvalidBindParameter            = errors.New("invalid bind parameter")
	ErrInvalidCollation                = errors.New("invalid collation")
	ErrInvalidInsertType               = errors.New("invalid insert type")
	ErrInvalidUpdateType               = errors.New("invalid update type")
	ErrInvalidSelectType               = errors.New("invalid select type")
	ErrInvalidJoinOperator             = errors.New("invalid join operator")
	ErrInvalidOrderType                = errors.New("invalid order type")
	ErrInvalidNullOrderType            = errors.New("invalid null order type")
	ErrInvalidReturningClause          = errors.New("invalid returning clause")
	ErrInvalidCompoundOperator         = errors.New("invalid compound operator")
	ErrInvalidUpsertType               = errors.New("invalid upsert type")
	ErrInvalidUnaryOperator            = errors.New("invalid unary operator")
	ErrInvalidBinaryOperator           = errors.New("invalid binary operator")
	ErrInvalidStringComparisonOperator = errors.New("invalid string comparison operator")
	ErrInvalidArithmeticOperator       = errors.New("invalid arithmetic operator")
	ErrUnknownTable                    = errors.New("unknown table")
)

// wrapErr wraps an error with another, if the second error is not nil
func wrapErr(err error, err2 error) error {
	if err2 == nil {
		return nil
	}
	return fmt.Errorf("%w: %s", err, err2)
}
