package types

import (
	"errors"
)

var ErrNotEnoughFunds = errors.New("not enough funds")
var ErrFeeTooLow = errors.New("the sent fee is too low for the requested operation")
var ErrInvalidSignature = errors.New("invalid signature")
var ErrInvalidID = errors.New("could not reconstruct the provided ID")
var ErrIncorrectOperation = errors.New("incorrect operation")
var ErrIncorrectCrud = errors.New("incorrect crud")
