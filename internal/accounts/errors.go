package accounts

import (
	"errors"
	"fmt"
	"math/big"
)

// These are common error values returned from account methods. They may be used
// with errors.Is for less error prone coding of the error cause, while
// SpendError may be used with errors.As to get pre-Spend account details if it
// was available.
var (
	ErrInsufficientFunds = fmt.Errorf("insufficient funds")
	ErrConvertToBigInt   = fmt.Errorf("could not convert to big int")
	ErrInvalidNonce      = fmt.Errorf("invalid nonce")
	ErrAccountNotFound   = fmt.Errorf("account not found")
)

// SpendError is used to inform the caller of the account's starting balance and
// nonce in certain spend errors. This is helpful on paths where an sufficient
// balance zeros the account balance. Only a pointer implements the error
// interface, and as such using errors.As is only possible via `se :=
// new(SpendError)` and `errors.As(err, &se)` (other uses are rejected at
// compile time to avoid gotchas).
type SpendError struct {
	Err     error
	Balance *big.Int
	Nonce   int64
}

// NewSpendError constructs a new SpendError from a cause error and the accounts
// starting balance and nonce.
func NewSpendError(err error, bal *big.Int, nonce int64) *SpendError {
	if err == nil {
		err = errors.New("unknown spend error")
	}
	return &SpendError{
		Err:     err,
		Balance: bal,
		Nonce:   nonce,
	}
}

// Error satisfies the error interface. We make this a pointer receiver to avoid
// pitfalls trying to use values As pointers and vice versa.
func (se *SpendError) Error() string {
	return fmt.Sprintf("%v: account balance %v, nonce %d", se.Err.Error(), se.Balance, se.Nonce)
}

// Unwrap allows errors.Is to identify wrapped errors.
func (se *SpendError) Unwrap() error {
	return se.Err
}

var _ error = (*SpendError)(nil)
