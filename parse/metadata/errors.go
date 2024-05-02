package metadata

import (
	"errors"
)

var (
	ErrReadOnlyProcedureContainsDML   = errors.New("read-only procedure contains DML statement")
	ErrReadOnlyProcedureCallsMutative = errors.New("read-only procedure calls mutative procedure")
	ErrUndeclaredVariable             = errors.New("undeclared variable")
	ErrVariableAlreadyDeclared        = errors.New("variable already declared")
	ErrIncorrectAssignmentType        = errors.New("incorrect assignment type")
	ErrAssignmentTypeMismatch         = errors.New("assignment type does not match variable type")
	ErrUntypedVariable                = errors.New("untyped variable")
	ErrUnknownContextualVariable      = errors.New("unknown contextual variable")
	ErrUnknownFunctionOrProcedure     = errors.New("unknown procedure/function")
	ErrUnknownForeignProcedure        = errors.New("unknown foreign procedure")
	ErrUnknownField                   = errors.New("unknown field")
	ErrIncorrectReturnType            = errors.New("incorrect return type")
	ErrComparisonTypesDoNotMatch      = errors.New("comparison types do not match")
	ErrArrayElementTypesDoNotMatch    = errors.New("array element type does not match the expected type")
	ErrBreakUsedOutsideOfLoop         = errors.New("BREAK cannot be used outside of a loop")
	ErrReturnNextUsedOutsideOfLoop    = errors.New("RETURN NEXT cannot be used outside of a loop")
	ErrReturnNextUsedInNonTableProc   = errors.New("RETURN NEXT cannot be used in a procedure that does not return a table")
	ErrReturnNextInvalidCount         = errors.New("RETURN NEXT must return the same number of fields as the procedure return")
	ErrForeignCallMissingField        = errors.New("missing field in foreign call")
)
