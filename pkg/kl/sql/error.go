package sql

import "errors"

var (
	ErrSyntax                      = errors.New("syntax error")
	ErrFunctionNotSupported        = errors.New("function not supported")
	ErrKeywordNotSupported         = errors.New("keyword not supported")
	ErrSelectFromMultipleTables    = errors.New("implicit cartesian join(1) is not supported")
	ErrJoinWithoutCondition        = errors.New("implicit cartesian join(2) is not supported")
	ErrJoinWithTrueCondition       = errors.New("implicit cartesian join(3) is not supported")
	ErrJoinUsingNotSupported       = errors.New("join using is not supported")
	ErrJoinConditionOpNotSupported = errors.New("join condition operator is not supported")
	ErrJoinNotSupported            = errors.New("join type is not supported")
	ErrMultiJoinNotSupported       = errors.New("multi joins are not supported")
	ErrBindParameterNotFound       = errors.New("bind parameter not found")
)
