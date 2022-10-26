package wallet

import "kwil/x/async"

// RequestEvents
// Consumed by processor
type RequestEvents interface {
	OnWithdrawal(func(WithdrawalEvent) async.Action) // topic: responses, type: withdrawal
	OnSpend(func(SpendEvent) async.Action)           // topic: responses, type: spend
}

type WithdrawalEvent struct {
	// ...
}

type SpendEvent struct {
	// ...
}
