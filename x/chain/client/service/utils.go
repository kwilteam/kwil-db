package service

import (
	"kwil/x/chain/types"

	"github.com/ethereum/go-ethereum/ethclient"
)

// This isn't best practice since these are simply passthroughs to the provider

func (c *chainClient) ChainCode() types.ChainCode {
	return c.provider.ChainCode()
}

func (c *chainClient) AsEthClient() (*ethclient.Client, error) {
	return c.provider.AsEthClient()
}
