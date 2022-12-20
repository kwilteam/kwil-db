package client

import (
	"math/big"

	ethc "github.com/ethereum/go-ethereum/ethclient"
)

type EVMClient struct {
	client  *ethc.Client
	chainId *big.Int
}

func New(client *ethc.Client, chainId *big.Int) *EVMClient {
	return &EVMClient{
		client:  client,
		chainId: chainId,
	}
}
