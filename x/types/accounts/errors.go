package accounts

import "errors"

var (
	ErrAccountNotRegistered = errors.New("config not registered")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidNonce         = errors.New("invalid nonce")
)
