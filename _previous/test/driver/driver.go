// Package driver contains concrete drivers for tests.
//

package driver

import (
	"errors"
)

var ErrTxNotConfirmed = errors.New("transaction not confirmed")
