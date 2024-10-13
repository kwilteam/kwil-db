package common

import "errors"

var (
	ErrTypeMismatch     = errors.New("type mismatch")
	ErrIndexOutOfBounds = errors.New("index out of bounds")
)
