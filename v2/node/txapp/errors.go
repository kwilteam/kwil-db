package txapp

import "errors"

var (
	ErrCallerNotValidator = errors.New("caller is not a validator")
	ErrCallerIsValidator  = errors.New("caller is already a validator")
	ErrCallerNotProposer  = errors.New("caller is not the block proposer")
	ErrTargetNotValidator = errors.New("target is not a validator")
)
