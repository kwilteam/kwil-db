package balances

import "fmt"

var (
	ErrInsufficientFunds = fmt.Errorf("insufficient funds")
	ErrAccountNotFound   = fmt.Errorf("account not found")
	ErrConvertToBigInt   = fmt.Errorf("could not convert to big int")
	ErrInvalidNonce      = fmt.Errorf("invalid nonce")
)
