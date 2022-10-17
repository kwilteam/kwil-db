package pub

import (
	"context"
	"fmt"
	"kwil/x"
	"kwil/x/async"
)

var ErrProducerClosed = fmt.Errorf("emitter closed")

type Emitter[T any] interface {
	ID() int

	// Send publishes a message to an underlying message
	// provider. An Action will be returned to asynchronously
	// signal the message succeeded or failed to publish.
	// The default topic for the Emitter will be used. If no
	// default has been configured, then an error will be returned.
	Send(ctx context.Context, item T) async.Action

	// SendT publishes a message to an underlying message
	// provider. An Action will be returned to asynchronously
	// signal the message succeeded or failed to publish. The
	// provided topic will be used in place of a configured
	// default.
	SendT(ctx context.Context, topic string, item T) async.Action

	// SendSync delegates to Send and blocks until the returned
	// Action is completed.
	SendSync(ctx context.Context, item T) error

	// SendSyncT delegates to SendT and blocks until the returned
	// Action is completed.
	SendSyncT(ctx context.Context, topic string, item T) error

	// SendNoAck publishes a message to an underlying message
	// provider. An error will only be returned of the message
	// was un able to be enqueued (e.g., due to closure, etc.).
	// No ack will be made and no guarantees that the message
	// was successfully delivered can be made.
	SendNoAck(ctx context.Context, item T) error

	// SendNoAckT publishes a message with the topic specified
	// to an underlying message provider. An error will only
	// be returned of the message was un able to be enqueued
	// (e.g., due to closure, etc.). No ack will be made and
	// no guarantees that the message was successfully delivered
	// can be made.
	SendNoAckT(ctx context.Context, topic string, item T) error

	// Close closes the emitter and releases all resources.
	Close()

	// CloseAndWait closes the emitter and releases all resources.
	CloseAndWait(ctx context.Context) error

	// OnClosed returns a channel that is closed when the
	// emitter is closed.
	OnClosed() <-chan x.Void
}
