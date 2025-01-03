package interpreter

import (
	"errors"
	"fmt"
)

var (
	ErrUnaryOnNonScalar      = errors.New("cannot perform unary operation on a non-scalar value")
	ErrTypeMismatch          = errors.New("type mismatch")
	ErrIndexOutOfBounds      = errors.New("index out of bounds")
	ErrVariableNotFound      = errors.New("variable not found")
	ErrStatementMutatesState = errors.New("statement mutates state")
	ErrActionMutatesState    = errors.New("action mutates state")
	ErrActionOwnerOnly       = errors.New("action is owner-only")
	ErrActionPrivate         = errors.New("action is private")
	ErrSystemOnly            = errors.New("system-only action")
	ErrCannotDrop            = errors.New("cannot drop")
	ErrCannotCall            = errors.New("cannot call action")
	ErrDoesNotHavePriv       = errors.New("user does not have privilege")
	ErrNamespaceNotFound     = errors.New("namespace not found")
	ErrNamespaceExists       = errors.New("namespace already exists")
	ErrArithmetic            = errors.New("arithmetic error")
	ErrComparison            = errors.New("comparison error")
	ErrCast                  = errors.New("type cast error")
	ErrUnary                 = errors.New("unary operation error")
	ErrArrayMixedTypes       = errors.New("array contains mixed types")
)

func castErr(e error) error {
	return fmt.Errorf("%w: %s", ErrCast, e)
}
