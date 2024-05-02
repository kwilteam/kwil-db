package types

import (
	"errors"
)

var (
	// ErrSyntaxError is returned when a syntax error is encountered.
	ErrSyntaxError = errors.New("syntax error")

	ErrReadOnlyProcedureContainsDML   = errors.New("read-only procedure contains DML statement")
	ErrReadOnlyProcedureCallsMutative = errors.New("read-only procedure calls mutative procedure")
	ErrUndeclaredVariable             = errors.New("undeclared variable")
	ErrVariableAlreadyDeclared        = errors.New("variable already declared")
	ErrUntypedVariable                = errors.New("untyped variable")
	ErrUnknownContextualVariable      = errors.New("unknown contextual variable")
	ErrUnknownFunctionOrProcedure     = errors.New("unknown procedure/function")
	ErrUnknownForeignProcedure        = errors.New("unknown foreign procedure")
	ErrUnknownField                   = errors.New("unknown field")
	ErrArgCount                       = errors.New("argument count mismatch")

	// unknown reference errors
	ErrForeignCallMissingField = errors.New("missing field in foreign call")

	// Type errors
	ErrNotNumericType = errors.New("not a numeric type")
	ErrReturnCount    = errors.New("invalid number of return values") // used forr invalid number of returns
	ErrAssignment     = errors.New("assignment error")
	ErrComparisonType = errors.New("comparison types do not match")
	ErrArrayType      = errors.New("incorrect type for array")
	ErrArithmeticType = errors.New("arithmetic types do not match")
	ErrParamType      = errors.New("variable type does not match function parameter type")

	// procedure return errors
	ErrReturnNextUsedInNonTableProc = errors.New("RETURN NEXT cannot be used in a procedure that does not return a table")
	ErrReturnNextInvalidCount       = errors.New("RETURN NEXT must return the same number of fields as the procedure return")

	// Loop errors
	ErrInvalidIterable             = errors.New("invalid iterable")
	ErrBreakUsedOutsideOfLoop      = errors.New("BREAK cannot be used outside of a loop")
	ErrReturnNextUsedOutsideOfLoop = errors.New("RETURN NEXT cannot be used outside of a loop")
)
