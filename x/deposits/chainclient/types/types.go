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
	GetLatestBlock(context.Context) (int64, error)
	GetContract(string) (Contract, error)
}

type Contract interface {
	GetDeposits(context.Context, int64, int64) ([]*Deposit, error)
}

type Log interface {
	Type() uint8
	Height() int64
	Tx() string
}

type IDeposit interface {
	Log
	Token() string
	Amount() string
	Target() string
	Caller() string
}

type IWithdrawal interface {
	Log
	Token() string
	Amount() string
	Recipient() string
	Sender() string
}

type Deposit struct {
	caller string
	target string
	amount string
	height int64
	tx     string
	typ    uint8
	token  string
}

func NewDeposit(caller, target, amount string, height int64, tx string, typ uint8, token string) *Deposit {
	return &Deposit{
		caller: caller,
		target: target,
		amount: amount,
		height: height,
		tx:     tx,
		typ:    typ,
		token:  token,
	}
}

func (d *Deposit) Type() uint8 {
	return d.typ
}

func (d *Deposit) Height() int64 {
	return d.height
}

func (d *Deposit) Tx() string {
	return d.tx
}

func (d *Deposit) Token() string {
	return d.token
}

func (d *Deposit) Amount() string {
	return d.amount
}

func (d *Deposit) Target() string {
	return d.target
}

func (d *Deposit) Caller() string {
	return d.caller
}
