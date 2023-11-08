package contracts

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types/chain"
)

type TokenContract interface {
	Allowance(ctx context.Context, owner, spender string) (*big.Int, error)
	Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error)
	BalanceOf(ctx context.Context, address string) (*big.Int, error)
}

type EscrowContract interface {
	Deposit(ctx context.Context, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error)
	TokenAddress() string
	Balance(ctx context.Context, address string) (*big.Int, error)
	GetDeposits(ctx context.Context, from uint64, to *uint64) ([]*chain.DepositEvent, error)
}
