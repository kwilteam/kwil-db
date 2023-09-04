package balances

import (
	"fmt"
	"math/big"
)

// emptyAccount returns an empty account with a balance of 0 and a nonce of 0.
func emptyAccount() *Account {
	return &Account{
		PublicKey: nil,
		Balance:   big.NewInt(0),
		Nonce:     0,
	}
}

type Account struct {
	PublicKey []byte   `json:"public_key"`
	Balance   *big.Int `json:"balance"`
	Nonce     int64    `json:"nonce"`
}

// validateSpend validates that the account has enough funds to spend the amount.
// it returns the new balance, or an error if the account does not have enough funds.
func (a *Account) validateSpend(amount *big.Int) (*big.Int, error) {
	newBal := new(big.Int).Sub(a.Balance, amount)
	if newBal.Cmp(big.NewInt(0)) < 0 {
		return nil, ErrInsufficientFunds
	}
	return newBal, nil
}

// validateNonce checks that the passed nonce is exactly one greater than the account's nonce.
func (a *Account) validateNonce(nonce int64) error {
	if a.Nonce+1 != nonce {
		return fmt.Errorf("%w: expected %d, got %d", ErrInvalidNonce, a.Nonce+1, nonce)
	}
	return nil
}
