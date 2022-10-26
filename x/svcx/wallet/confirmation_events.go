package wallet

import (
	"kwil/x"
	"kwil/x/async"
)

// ConfirmationEvents
// All nodes consume
type ConfirmationEvents interface {
	OnDeposited(func(DepositedEvent) async.Action) // topic: confirmations, type: deposit
	OnWithdrawn(func(WithdrawnEvent) async.Action) // topic: confirmations, type: withdrawal
	OnSpent(func(SpentEvent) async.Action)         // topic: confirmations, type: spend

	Close() error
	OnClosed() <-chan x.Void
}
