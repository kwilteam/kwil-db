package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (c *contract) BalanceOf(address string) (*big.Int, error) {
	return c.ctr.BalanceOf(nil, common.HexToAddress(address))
}
