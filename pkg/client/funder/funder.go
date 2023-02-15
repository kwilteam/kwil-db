package funder

import (
	"context"
	escrowTypes "kwil/pkg/chain/contracts/escrow/types"
	tokenTypes "kwil/pkg/chain/contracts/token/types"
	"math/big"
)

type Funder interface {
	ApproveFunds(ctx context.Context, spender string, amount *big.Int) (*tokenTypes.ApproveResponse, error)
	DepositFunds(ctx context.Context, amount *big.Int) (*escrowTypes.DepositResponse, error)
}

type funder struct {
	providerAddress string
}

/*
func New(providerAddress string) Funder {
	return &funder{
		providerAddress: providerAddress,
	}
}
*/
