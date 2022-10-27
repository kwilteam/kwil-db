package wallet

import "kwil/x"

// RequestProcessor
// Background process enacted by topic events
// Leader elected singleton service
type RequestProcessor interface {
	// listens to request topic
	// writes to confirmation topic

	Start() error
	Stop() error
	OnStop() <-chan x.Void
}
