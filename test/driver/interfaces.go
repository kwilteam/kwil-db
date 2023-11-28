package driver

import (
	"context"
	"crypto/ecdsa"
	"math/big"
)

type BridgeClient interface {
	Allowance(ctx context.Context, owner, spender string) (*big.Int, error)
	Approve(ctx context.Context, spender string, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error)
	BalanceOf(ctx context.Context, address string) (*big.Int, error)
	Deposit(ctx context.Context, amount *big.Int, privateKey *ecdsa.PrivateKey) (string, error)
	TokenAddress() string
	DepositBalance(ctx context.Context, address string) (*big.Int, error)
}
