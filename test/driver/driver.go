// Package driver contains concrete drivers for tests.
//

package driver

import (
	"errors"
)

var (
	ErrTxNotConfirmed = errors.New("transaction not confirmed")

	// ErrTxNotFound is returned if a requested transaction is just not found.
	// Unfortunately this happens frequently for a very brief period after
	// broadcast. The bug is of unknown cause, but it seems like cometbft since
	// query-tx is asking cometbft and we use "sync" broadcast meaning it was
	// accepted into mempool. If it were mined then we also would have found it.
	ErrTxNotFound = errors.New("transaction NOT FOUND")
)
