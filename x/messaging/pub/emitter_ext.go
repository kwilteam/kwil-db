package pub

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"kwil/x"
	"kwil/x/messaging/mx"
	"kwil/x/syncx"
)

type emitter[T any] struct {
	id     int
	client *emitter_client
	serdes mx.Serdes[T]
	fn     func() <-chan x.Void
	done   syncx.Chan[x.Void]
}

func (e *emitter[T]) Send(ctx context.Context, message *Message[T]) error {
	m, err := e.createMessage(message)
	if err != nil {
		return err
	}

	return e.client.send_async(ctx, m, message.AckNack)
}

func (e *emitter[T]) ID() int {
	return e.id
}

func (e *emitter[T]) Close() {
	e.fn()
	e.done.Close()
}

func (e *emitter[T]) CloseAndWait(ctx context.Context) error {
	e.Close()
	select {
	case <-e.done.ClosedCh():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *emitter[T]) OnClosed() <-chan x.Void {
	return e.done.ClosedCh()
}

func (e *emitter[T]) createMessage(message *Message[T]) (*kgo.Record, error) {
	key, payload, err := e.serdes.Serialize(message.Payload)
	if err != nil {
		return nil, err
	}

	return &kgo.Record{
		Key:   key,
		Value: payload,
		Topic: message.Topic,
	}, nil
}
