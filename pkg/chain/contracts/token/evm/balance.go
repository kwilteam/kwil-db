package evm

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	return c.ctr.BalanceOf(&bind.CallOpts{
		Context: ctx,
	}, common.HexToAddress(address))
}
