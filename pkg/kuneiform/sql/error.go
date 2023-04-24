package sql

import (
	"errors"
	"fmt"
)

const (
	JoinCountAllowed = 3 // 4 tables
)

var (
	ErrSyntax                          = errors.New("syntax error")
	ErrTableNotFound                   = errors.New("table not found")
	ErrColumnNotFound                  = errors.New("column not found")
	ErrCreateTableNotSupported         = errors.New("create table is not supported")
	ErrCreateIndexNotSupported         = errors.New("create index is not supported")
	ErrCreateViewNotSupported          = errors.New("create view is not supported")
	ErrCreateTriggerNotSupported       = errors.New("create trigger is not supported")
	ErrCreateVirtualTableNotSupported  = errors.New("create virtual table is not supported")
	ErrDropTableNotSupported           = errors.New("drop statement is not supported")
	ErrAlterTableNotSupported          = errors.New("alter table is not supported")
	ErrFunctionNotSupported            = errors.New("function not supported")
	ErrKeywordNotSupported             = errors.New("keyword not supported")
	ErrSelectFromMultipleTables        = errors.New("implicit cartesian join(1) is not supported")
	ErrJoinWithoutCondition            = errors.New("implicit cartesian join(2) is not supported")
	ErrJoinWithTrueCondition           = errors.New("implicit cartesian join(3) is not supported")
	ErrJoinUsingNotSupported           = errors.New("join using is not supported")
	ErrJoinConditionOpNotSupported     = errors.New("join condition operator is not supported")
	ErrJoinConditionFuncNotSupported   = errors.New("join condition on function is not supported")
	ErrJoinConditionTooDeep            = errors.New("join condition is too deep")
	ErrJoinConditionNotSupported       = errors.New("join condition is not supported")
	ErrJoinNotSupported                = errors.New("join type is not supported")
	ErrMultiJoinNotSupported           = fmt.Errorf("multi joins(>%d) are not supported", JoinCountAllowed)
	ErrBindParameterNotFound           = errors.New("bind parameter not found")
	ErrBindParameterPrefixNotSupported = errors.New("bind parameter prefix not supported")
	ErrModifierNotSupported            = errors.New("modifier not supported")
)
