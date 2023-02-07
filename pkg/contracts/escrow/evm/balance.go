package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"kwil/pkg/types/contracts/escrow"
)

func (c *contract) Balance(ctx context.Context, params *escrow.DepositBalanceParams) (*escrow.DepositBalanceResponse, error) {
	cAuth := &bind.CallOpts{
		Pending: true,
		From:    common.HexToAddress(params.Validator),
		Context: ctx,
	}
	balance, err := c.ctr.Balance(cAuth, common.HexToAddress(params.Validator), common.HexToAddress(params.Address))
	if err != nil {
		return nil, err
	}

	return &escrow.DepositBalanceResponse{Balance: balance}, nil
}
