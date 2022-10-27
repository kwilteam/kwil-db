package wallet

import (
	"kwil/x"
	"kwil/x/async"
)

// ConfirmationEvents background process consuming
// events emitted to confirmation topic by
// Consumes all topic partitions for events per node
type ConfirmationEvents interface {
	OnDeposited(func(DepositedEvent) async.Action)
	OnWithdrawn(func(WithdrawnEvent) async.Action)
	OnSpent(func(SpentEvent) async.Action)

	Start() error
	Stop() error

	OnStop() <-chan x.Void
}
