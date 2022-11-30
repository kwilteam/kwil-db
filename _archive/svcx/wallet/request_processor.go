package wallet

import (
	"kwil/archive/svcx/messaging/mx"
	"kwil/x"
	"kwil/x/async"
)

type MessageTransform func(*mx.RawMessage) async.Task[*mx.RawMessage]

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
