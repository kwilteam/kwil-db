package fund

import (
	"context"
	"crypto/ecdsa"
	"kwil/x/types/contracts/escrow"
	"kwil/x/types/contracts/token"
	"math/big"
)

type IFund interface {
	ApproveToken(ctx context.Context, pk *ecdsa.PrivateKey, spender string, amount *big.Int) (*token.ApproveResponse, error)
	GetBalance(ctx context.Context, account string) (*big.Int, error)
	GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error)
	DepositFund(ctx context.Context, pk *ecdsa.PrivateKey, to string, amount *big.Int) (*escrow.DepositResponse, error)
	GetDepositBalance(ctx context.Context, validator string, wallet string) (*big.Int, error)
	GetConfig() *Config
}
