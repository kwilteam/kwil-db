package events

import (
	"math/big"
)

type Event interface {
	GetName() string
	GetHeight() *big.Int
	GetData() interface{}
	GetTx() []byte
}

type DepositEvent struct {
	Name   string
	Height *big.Int
	Data   *Deposit
	Tx     []byte
}

func (e *DepositEvent) GetName() string {
	return e.Name
}

func (e *DepositEvent) GetHeight() *big.Int {
	return e.Height
}

func (e *DepositEvent) GetData() interface{} {
	return e.Data
}

func (e *DepositEvent) GetTx() []byte {
	return e.Tx
}
