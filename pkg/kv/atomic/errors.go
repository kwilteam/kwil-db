package atomic

import "errors"

var (
	ErrSessionActive    = errors.New("session already active")
	ErrSessionNotActive = errors.New("session not active")
	ErrTxnActive        = errors.New("transaction already active")
	ErrTxnNotActive     = errors.New("transaction not active")
)
