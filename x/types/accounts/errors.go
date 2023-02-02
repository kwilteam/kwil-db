package accounts

import "errors"

var (
	ErrAccountNotRegistered = errors.New("info not registered")
	ErrInsufficientFunds    = errors.New("insufficient funds")
	ErrInvalidNonce         = errors.New("invalid nonce")
)
