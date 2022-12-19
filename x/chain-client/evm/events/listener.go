package events

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	ethc "github.com/ethereum/go-ethereum/ethclient"
)

// implements the chain-client/service/Subscriber interface
type EVMEventListener struct {
	client  *ethc.Client
	chainId *big.Int
}

type EVMSubscription struct {
	blocks chan int64
	errs   chan error
	sub    ethereum.Subscription
}

func New(endpoint string, chainId *big.Int) (*EVMEventListener, error) {
	client, err := ethc.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	return &EVMEventListener{
		client:  client,
		chainId: chainId,
	}, nil
}

func newSub(sub ethereum.Subscription) *EVMSubscription {
	return &EVMSubscription{
		blocks: make(chan int64),
		errs:   make(chan error),
		sub:    sub,
	}
}
