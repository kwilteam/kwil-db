package interpreter

import "errors"

var (
	ErrUnaryOnNonScalar   = errors.New("cannot perform unary operation on a non-scalar value")
	ErrTypeMismatch       = errors.New("type mismatch")
	ErrIndexOutOfBounds   = errors.New("index out of bounds")
	ErrVariableNotFound   = errors.New("variable not found")
	ErrActionMutatesState = errors.New("action mutates state")
	ErrActionOwnerOnly    = errors.New("action is owner-only")
)
