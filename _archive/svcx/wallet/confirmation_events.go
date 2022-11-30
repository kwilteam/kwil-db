package wallet

import (
	"kwil/x"
)

// ConfirmationEvents background process consuming
// events emitted to confirmation topic by
// Consumes all topic partitions for events per node
type ConfirmationEvents interface {
	Start() error
	Stop() error
	OnStop() <-chan x.Void
}
