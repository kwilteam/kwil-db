package client

import "github.com/ethereum/go-ethereum"

type EVMSubscription struct {
	blocks chan int64
	errs   chan error
	sub    ethereum.Subscription
}

func (e *EVMSubscription) Unsubscribe() {
	e.sub.Unsubscribe()
}

func (e *EVMSubscription) Err() <-chan error {
	return e.errs
}

func (e *EVMSubscription) Blocks() <-chan int64 {
	return e.blocks
}
