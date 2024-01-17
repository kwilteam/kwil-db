package txapp

import "errors"

var (
	ErrCallerNotValidator = errors.New("caller is not a validator")
	ErrCallerNotProposer  = errors.New("caller is not the block proposer")
)
