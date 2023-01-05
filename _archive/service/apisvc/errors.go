package apisvc

import (
	"errors"
)

var (
	ErrNotEnoughFunds     = errors.New("not enough funds")
	ErrFeeTooLow          = errors.New("the sent fee is too low for the requested operation")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrInvalidID          = errors.New("could not reconstruct the provided ID")
	ErrIncorrectOperation = errors.New("incorrect operation")
	ErrIncorrectCrud      = errors.New("incorrect crud")
	ErrInvalidDDLType     = errors.New("invalid ddl type")
	ErrInvalidAmount      = errors.New("invalid amount")
)
