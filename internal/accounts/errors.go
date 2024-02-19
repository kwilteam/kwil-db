package accounts

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

var (
	ErrInsufficientFunds = fmt.Errorf("insufficient funds")
	ErrConvertToBigInt   = fmt.Errorf("could not convert to big int")
	ErrInvalidNonce      = fmt.Errorf("invalid nonce")
	ErrAccountNotFound   = fmt.Errorf("account not found")
	ErrNegativeBalance   = fmt.Errorf("negative balance not permitted")
	ErrNegativeTransfer  = fmt.Errorf("negative transfer not permitted")
)

// errInsufficientFunds formats an error message for insufficient funds
func errInsufficientFunds(account []byte, amount *big.Int, balance *big.Int) error {
	return fmt.Errorf("%w: account %s tried to use %s, but only has balance %s", ErrInsufficientFunds, hex.EncodeToString(account), amount.String(), balance.String())
}
