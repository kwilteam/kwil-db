package types

import (
	"context"
	"crypto/ecdsa"
	"math/big"
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
	GetDeposits(context.Context, int64, int64, string) ([]*Deposit, error)
	GetWithdrawals(context.Context, int64, int64, string) ([]*Withdrawal, error)
	ReturnFunds(context.Context, *ecdsa.PrivateKey, string, string, *big.Int, *big.Int) error
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

type Withdrawal struct {
	caller   string
	receiver string
	amount   string
	fee      string
	nonce    string
	height   int64
	tx       string
	typ      uint8
	token    string
}

func NewWithdrawal(caller, receiver, amount, fee, nonce string, height int64, tx string, typ uint8, token string) *Withdrawal {
	return &Withdrawal{
		caller:   caller,
		receiver: receiver,
		amount:   amount,
		fee:      fee,
		nonce:    nonce,
		height:   height,
		tx:       tx,
		typ:      typ,
		token:    token,
	}
}

func (w *Withdrawal) Type() uint8 {
	return w.typ
}

func (w *Withdrawal) Height() int64 {
	return w.height
}

func (w *Withdrawal) Tx() string {
	return w.tx
}

func (w *Withdrawal) Token() string {
	return w.token
}

func (w *Withdrawal) Amount() string {
	return w.amount
}

func (w *Withdrawal) Fee() string {
	return w.fee
}

func (w *Withdrawal) Nonce() string {
	return w.nonce
}

func (w *Withdrawal) Receiver() string {
	return w.receiver
}

func (w *Withdrawal) Caller() string {
	return w.caller
}
