package interpreter

import "errors"

var (
	ErrUnaryOnNonScalar = errors.New("cannot perform unary operation on a non-scalar value")

	ErrVariableNotFound = errors.New("variable not found")
)
