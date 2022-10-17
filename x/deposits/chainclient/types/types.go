package types

import (
	"context"
)

type BlockSubscription interface {
	Unsubscribe()

	Err() <-chan error
}

type Client interface {
	SubscribeBlocks(context.Context, chan<- int64) (BlockSubscription, error)
	GetContract(string) (Contract, error)
}

type Contract interface {
	GetDeposits(context.Context, int64, int64) ([]Deposit, error)
}

type Log interface {
	Type() uint8
	Height() int64
	Tx() string
}

type Deposit interface {
	Log
	Token() string
	Amount() string
	Target() string
	Caller() string
}

type Withdrawal interface {
	Log
	Token() string
	Amount() string
	Recipient() string
	Sender() string
}
