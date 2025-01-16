package engine

import "errors"

var (
	// Errors that suggest a bug in a user's executing code. These are type errors,
	// issues in arithmetic, array indexing, etc.
	ErrType                = errors.New("type error")
	ErrReturnShape         = errors.New("unexpected action/function return shape")
	ErrUnknownVariable     = errors.New("unknown variable")
	ErrInvalidVariable     = errors.New("invalid variable name")
	ErrLoop                = errors.New("loop error")
	ErrArithmetic          = errors.New("arithmetic error")
	ErrComparison          = errors.New("comparison error")
	ErrCast                = errors.New("type cast error")
	ErrUnary               = errors.New("unary operation error")
	ErrIndexOutOfBounds    = errors.New("index out of bounds")
	ErrArrayDimensionality = errors.New("array dimensionality error")
	ErrInvalidNull         = errors.New("invalid null value")
	ErrArrayTooSmall       = errors.New("array too small")
	ErrExtensionInvocation = errors.New("extension invocation error")

	// Errors that signal the existence or non-existence of an object.
	ErrUnknownAction     = errors.New("unknown action")
	ErrUnknownTable      = errors.New("unknown table")
	ErrNamespaceNotFound = errors.New("namespace not found")
	ErrNamespaceExists   = errors.New("namespace already exists")

	// Errors that likely are not the result of a user error, but instead are informing
	// the user of an operation that is not allowed in order to maintain the integrity of
	// the system.
	ErrCannotMutateState          = errors.New("connection is read-only and cannot mutate state")
	ErrIllegalFunctionUsage       = errors.New("illegal function usage")
	ErrQueryActive                = errors.New("a query is currently active. nested queries are not allowed")
	ErrCannotBeNamespaced         = errors.New("the selected object is global-only, and cannot be namespaced")
	ErrCannotMutateExtension      = errors.New("cannot mutate an extension's schema or data directly")
	ErrCannotMutateInfoNamespace  = errors.New(`cannot mutate the "info" namespace directly`)
	ErrCannotDropBuiltinNamespace = errors.New("cannot drop a built-in namespace")
	ErrBuiltInRole                = errors.New("invalid operation on built-in role")
	ErrInvalidTxCtx               = errors.New("invalid transaction context")

	// Errors that are the result of not having proper permissions or failing to meet a condition
	// that was programmed by the user.
	ErrActionOwnerOnly      = errors.New("action is owner-only")
	ErrActionPrivate        = errors.New("action is private")
	ErrActionSystemOnly     = errors.New("action is system-only")
	ErrDoesNotHavePrivilege = errors.New("user does not have privilege")

	// Errors that signal an error in a deeper layer, and originate from
	// somewhere deeper than the interpreter.
	ErrParse        = errors.New("parse error")
	ErrQueryPlanner = errors.New("query planner error")
	ErrPGGen        = errors.New("postgres SQL generation error")
)
