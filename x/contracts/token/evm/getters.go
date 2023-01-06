package evm

import "math/big"

// getters that are set at contract creation

func (c *contract) Address() string {
	return c.address
}

func (c *contract) Name() string {
	return c.tokenName
}

func (c *contract) Symbol() string {
	return c.tokenSymbol
}

func (c *contract) Decimals() uint8 {
	return c.decimals
}

func (c *contract) TotalSupply() *big.Int {
	return c.totalSupply
}
