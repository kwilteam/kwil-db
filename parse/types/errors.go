package types

import (
	"errors"
)

var (
	ErrReadOnlyProcedureContainsDML   = errors.New("read-only procedure contains DML statement")
	ErrReadOnlyProcedureCallsMutative = errors.New("read-only procedure calls mutative procedure")

	ErrUntypedVariable           = errors.New("untyped variable")
	ErrUnknownContextualVariable = errors.New("unknown contextual variable")
	ErrUnknownForeignProcedure   = errors.New("unknown foreign procedure")
	ErrUnknownField              = errors.New("unknown field")
	ErrArgCount                  = errors.New("argument count mismatch")

	// unknown reference errors
	ErrForeignCallMissingField = errors.New("missing field in foreign call")

	// Type errors
	ErrNotNumericType = errors.New("not a numeric type")
	ErrReturnCount    = errors.New("invalid number of return values") // used forr invalid number of returns
	ErrComparisonType = errors.New("comparison types do not match")
	ErrArrayType      = errors.New("incorrect type for array")
	ErrArithmeticType = errors.New("arithmetic types do not match")
	ErrParamType      = errors.New("variable type does not match function parameter type")

	// procedure return errors
	ErrReturnNextUsedInNonTableProc = errors.New("RETURN NEXT cannot be used in a procedure that does not return a table")
	ErrReturnNextInvalidCount       = errors.New("RETURN NEXT must return the same number of fields as the procedure return")

	// Loop errors

	ErrBreakUsedOutsideOfLoop      = errors.New("BREAK cannot be used outside of a loop")
	ErrReturnNextUsedOutsideOfLoop = errors.New("RETURN NEXT cannot be used outside of a loop")
)

var (
	// TODO: these errors are more general, and we should delete the above ones once we have refactored the code.
	// ErrSyntaxError is returned when a syntax error is encountered.
	ErrSyntaxError                = errors.New("syntax error")
	ErrInvalidIterable            = errors.New("invalid iterable")
	ErrUndeclaredVariable         = errors.New("undeclared variable")
	ErrVariableAlreadyDeclared    = errors.New("variable already declared")
	ErrType                       = errors.New("type error")
	ErrAssignment                 = errors.New("assignment error")
	ErrUnknownTable               = errors.New("unknown table reference")
	ErrTableDefinition            = errors.New("table definition error")
	ErrUnknownColumn              = errors.New("unknown column reference")
	ErrAmbiguousColumn            = errors.New("ambiguous column reference")
	ErrUnknownFunctionOrProcedure = errors.New("unknown function or procedure")
	// ErrFunctionSignature is returned when a function/procedure is called with the wrong number of arguments,
	// or returns an unexpected number of values / table.
	ErrFunctionSignature  = errors.New("function/procedure signature error")
	ErrTableAlreadyExists = errors.New("table already exists")
	// ErrResultShape is used if the result of a query is not in a shape we expect.
	ErrResultShape         = errors.New("result shape error")
	ErrUnnamedResultColumn = errors.New("unnamed result column")
	ErrTableAlreadyJoined  = errors.New("table already joined")
	ErrUnnamedJoin         = errors.New("unnamed join")
	ErrJoin                = errors.New("join error")
	ErrBreak               = errors.New("break error")
	ErrReturn              = errors.New("return type error")
	ErrAggregate           = errors.New("aggregate error")
)
