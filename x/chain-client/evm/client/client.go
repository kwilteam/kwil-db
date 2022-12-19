package client

import (
	"kwil/x/chain-client/evm/events"
	"math/big"

	ethc "github.com/ethereum/go-ethereum/ethclient"
)

type EVMClient struct {
	client   *ethc.Client
	chainId  *big.Int
	listener *events.EVMEventListener
}

func New(endpoint string, chainId *big.Int) (*EVMClient, error) {
	client, err := ethc.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	return &EVMClient{
		client:  client,
		chainId: chainId,
	}, nil
}
