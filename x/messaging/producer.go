package messaging

import (
	"context"
	"fmt"
	"kwil/x"
	cfg "kwil/x/messaging/config"
	"kwil/x/syncx"
)

var ErrProducerClosed = fmt.Errorf("producer closed")
var ErrUnexpectedProducerError = fmt.Errorf("producer event response unknown")

type Producer[T any] interface {
	// Submit publishes a message to an underlying message
	// provider. If an ack is provided on the message, it will
	// be used to signal the message was successfully published, else
	// no status will be returned.
	Submit(ctx context.Context, message ProducerMessage[T]) error

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

	tp, kp, err := loadP(config)
	if err != nil {
		return nil, err
	}

	buffer := config.Int32("out_buffer_size", 10)
	if buffer < 0 {
		return nil, fmt.Errorf("out_buffer_size cannot be a negative #")
	}

	var out syncx.Chan[*messageWithCtx]
	if buffer == 0 {
		out = syncx.NewChan[*messageWithCtx]()
	} else {
		out = syncx.NewChanBuffered[*messageWithCtx](int(buffer))
	}

	p := &producer[T]{
		kp:     kp,
		topic:  tp,
		serdes: serdes,
		out:    out,
		done:   make(chan x.Void),
	}

	p.start()

	return p, nil
}
