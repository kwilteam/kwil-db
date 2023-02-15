package evm

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"kwil/pkg/chain/contracts/escrow/types"
)

func (c *contract) Balance(ctx context.Context, params *types.DepositBalanceParams) (*types.DepositBalanceResponse, error) {
	cAuth := &bind.CallOpts{
		Pending: true,
		From:    common.HexToAddress(params.Validator),
		Context: ctx,
	}
	balance, err := c.ctr.Balance(cAuth, common.HexToAddress(params.Address), common.HexToAddress(params.Validator))
	if err != nil {
		return nil, err
	}

	return &types.DepositBalanceResponse{Balance: balance}, nil
}
