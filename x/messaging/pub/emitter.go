package pub

import (
	"context"
	"fmt"
	"kwil/x"
)

var ErrProducerClosed = fmt.Errorf("emitter closed")
var ErrUnexpectedProducerError = fmt.Errorf("emitter event response unknown")

type Emitter[T any] interface {
	// Send publishes a message to an underlying message
	// provider. If an ack is provided on the message, it will
	// be used to signal the message was successfully published, else
	// no status will be returned.
	Send(ctx context.Context, message Message[T]) error

	// Close closes the emitter and releases all resources.
	Close()

	// CloseAndWait closes the emitter and releases all resources.
	CloseAndWait(ctx context.Context) error

	// OnClosed returns a channel that is closed when the
	// emitter is closed.
	OnClosed() <-chan x.Void
}
