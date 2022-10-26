package wallet

import "kwil/x"

// RequestProcessor
// Background process enacted by topic events
type RequestProcessor interface {
	// listens to request topic
	// writes to confirmation topic

	Start() error
	Stop() error
	OnStop() <-chan x.Void
}

// EthereumProcessor
// Background listens for ethereum withdrawal
// confirmations and wallet deposits
type EthereumProcessor interface {
	// listens to request topic
	// writes to confirmation topic

	Start() error
	Stop() error
	OnStop() <-chan x.Void
}
