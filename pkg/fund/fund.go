package fund

import (
	"context"
	"kwil/pkg/contracts/escrow/types"
	types2 "kwil/pkg/contracts/token/types"
	"math/big"
)

type IFund interface {
	ApproveToken(ctx context.Context, spender string, amount *big.Int) (*types2.ApproveResponse, error)
	GetBalance(ctx context.Context, account string) (*big.Int, error)
	GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error)
	DepositFund(ctx context.Context, to string, amount *big.Int) (*types.DepositResponse, error)
	GetDepositBalance(ctx context.Context, validator string) (*big.Int, error)
	Close() error
}
