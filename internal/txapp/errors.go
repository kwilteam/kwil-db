package txapp

import "errors"

var (
	ErrCallerNotValidator = errors.New("caller is not a validator")
)
