package accounts

import "errors"

var (
	ErrAccountNotRegistered = errors.New("account not registered")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidNonce         = errors.New("invalid nonce")
)
