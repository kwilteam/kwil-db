package sub

import (
	"context"
	"fmt"
	"kwil/x"
	"kwil/x/messaging/mx"
)

var ErrReceiverClosed = fmt.Errorf("receiver closed")
var ErrUnexpectedReceiverError = fmt.Errorf("receiver event response unknown")

type ReceiverChannel[T any] interface {
	Topic() string
	PartitionId() mx.PartitionId
	OnReceive() <-chan MessageIterator[T]

	OnClosed() <-chan x.Void

	Close()

	// CloseAndWait closes the emitter and releases all resources.
	CloseAndWait(ctx context.Context) error
}
