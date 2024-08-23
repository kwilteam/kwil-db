package common

import "errors"

var (
	ErrTypeMismatch            = errors.New("type mismatch")
	ErrNotArray                = errors.New("not an array")
	ErrArithmeticOnArray       = errors.New("cannot perform arithmetic operation on array")
	ErrIndexOutOfBounds        = errors.New("index out of bounds")
	ErrNegativeSubstringLength = errors.New("negative substring length not allowed")
)
