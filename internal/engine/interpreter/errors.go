package interpreter

import "errors"

var (
	ErrUnaryOnNonScalar      = errors.New("cannot perform unary operation on a non-scalar value")
	ErrTypeMismatch          = errors.New("type mismatch")
	ErrIndexOutOfBounds      = errors.New("index out of bounds")
	ErrVariableNotFound      = errors.New("variable not found")
	ErrStatementMutatesState = errors.New("statement mutates state")
	ErrActionMutatesState    = errors.New("action mutates state")
	ErrActionOwnerOnly       = errors.New("action is owner-only")
	ErrDoesNotHavePriv       = errors.New("does not have privilege")
	ErrNamespaceNotFound     = errors.New("namespace not found")
)
