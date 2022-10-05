package messaging

import (
	"context"
	"fmt"
	"kwil/x"
	cfg "kwil/x/messaging/config"
	"kwil/x/rx"
	"kwil/x/syncx"
)

var ErrProducerClosed = fmt.Errorf("producer closed")
var ErrUnexpectedProducerError = fmt.Errorf("producer event response unknown")

type Producer[T any] interface {
	// Submit publishes a message to an underlying message
	// provider.
	Submit(ctx context.Context, message T) rx.Continuation

	// Close closes the producer and releases all resources.
	Close()

	// OnClosed returns a channel that is closed when the
	// producer is closed.
	OnClosed() <-chan x.Void
}

func NewProducer[T any](config cfg.Config, serdes Serdes[T]) (Producer[T], error) {
	if serdes == nil {
		return nil, fmt.Errorf("serdes is nil")
	}

	tp, kp, err := load(config)
	if err != nil {
		return nil, err
	}

	p := &producer[T]{
		kp:     kp,
		topic:  tp,
		serdes: serdes,
		out:    syncx.NewChanBuffered[*messageWithCtx](10),
	}

	go p.beginEventProcessing()

	return p, nil
}
