package chain

import (
	"math/big"
)

type DepositStore interface {
	Deposit(amount *big.Int, addr string, tx []byte, height *big.Int) error
	GetBalance(addr string) (*big.Int, error)
	CommitBlock(height *big.Int) error
	GetLastHeight() (*big.Int, error)
	SetLastHeight(height *big.Int) error
}
