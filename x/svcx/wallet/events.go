package wallet

import "math/big"

type WithdrawnEvent struct {
	walletId   string
	balance    big.Int
	request_id string
	// ...
}

type SpentEvent struct {
	walletId   string
	balance    big.Int
	request_id string
	// ...
}

type DepositedEvent struct {
	walletId string
	balance  big.Int
	// ...
}

func (d *DepositedEvent) WalletId() string {
	return d.walletId
}

func (d *DepositedEvent) Current() big.Int {
	return d.balance
}

func (d *SpentEvent) WalletId() string {
	return d.walletId
}

func (d *SpentEvent) Current() big.Int {
	return d.balance
}

func (d *WithdrawnEvent) WalletId() string {
	return d.walletId
}

func (d *WithdrawnEvent) Current() big.Int {
	return d.balance
}
