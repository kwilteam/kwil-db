package metadata

import (
	"errors"
)

var (
	ErrReadOnlyProcedureContainsDML   = errors.New("read-only procedure contains DML statement")
	ErrReadOnlyProcedureCallsMutative = errors.New("read-only procedure calls mutative procedure")
	ErrUndeclaredVariable             = errors.New("undeclared variable")
	ErrUntypedVariable                = errors.New("untyped variable")
	ErrUnknownContextualVariable      = errors.New("unknown contextual variable")
	ErrUnknownFunctionOrProcedure     = errors.New("unknown procedure/function")
	ErrUnknownForeignProcedure        = errors.New("unknown foreign procedure")
)
