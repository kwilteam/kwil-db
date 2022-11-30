package pub

import (
	"context"
	"kwil/_archive/svcx/messaging/mx"
	"kwil/x"
	"kwil/x/async"

	"github.com/twmb/franz-go/pkg/kgo"
)

type emitter[T any] struct {
	id     int
	client *emitter_client
	serdes mx.Serdes[T]
	fn     func() bool
	done   chan x.Void
}

func (e *emitter[T]) ID() int {
	return e.id
}

func (e *emitter[T]) Send(ctx context.Context, item T) async.Action {
	return e.SendT(ctx, "", item)
}

func (e *emitter[T]) SendT(ctx context.Context, topic string, item T) async.Action {
	return e._send(ctx, topic, item)
}

func (e *emitter[T]) SendSync(ctx context.Context, item T) error {
	return e.SendSyncT(ctx, "", item)
}

func (e *emitter[T]) SendSyncT(ctx context.Context, topic string, item T) error {
	a := e._send(ctx, topic, item)
	<-a.DoneCh()
	return a.GetError()
}

func (e *emitter[T]) SendNoAck(ctx context.Context, item T) error {
	return e.SendNoAckT(ctx, "", item)
}

func (e *emitter[T]) SendNoAckT(ctx context.Context, topic string, item T) error {
	m, err := e.createMessageFrom(topic, item)
	if err != nil {
		return err
	}

	return e.client.send_async(ctx, m, nil)
}

func (e *emitter[T]) Close() {
	if e.fn() {
		close(e.done)
	}
}

func (e *emitter[T]) CloseAndWait(ctx context.Context) error {
	e.Close()
	select {
	case <-e.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *emitter[T]) OnClosed() <-chan x.Void {
	return e.done
}

func (e *emitter[T]) _send(ctx context.Context, topic string, item T) async.Action {
	m, err := e.createMessageFrom(topic, item)
	if err != nil {
		return async.FailedAction(err)
	}

	ack, a := ackAsync()
	err = e.client.send_async(ctx, m, ack)
	if err != nil {
		a.Fail(err)
	}

	return a
}

func (e *emitter[T]) createMessageFrom(topic string, item T) (*kgo.Record, error) {
	key, payload, err := e.serdes.Serialize(item)
	if err != nil {
		return nil, err
	}

	return &kgo.Record{
		Key:   key,
		Value: payload,
		Topic: topic,
	}, nil
}
