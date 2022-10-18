package store

import (
	"errors"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrTxExists          = errors.New("this transaction already exists, and needs to be committed")
	ErrInsufficientFunds = errors.New("not enough funds")
)
