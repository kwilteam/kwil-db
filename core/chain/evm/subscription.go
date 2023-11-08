package evm

import "github.com/ethereum/go-ethereum"

// implements dto.Subscription interface
type ethSubscription struct {
	errs chan error
	sub  ethereum.Subscription
}

func (e *ethSubscription) Unsubscribe() {
	e.sub.Unsubscribe()
}

func (e *ethSubscription) Err() <-chan error {
	return e.errs
}

func newEthSubscription(sub ethereum.Subscription) *ethSubscription {
	return &ethSubscription{
		errs: make(chan error),
		sub:  sub,
	}
}
