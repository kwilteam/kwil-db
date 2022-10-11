package pub

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"kwil/x"
	"kwil/x/messaging/mx"
	"kwil/x/syncx"
	"kwil/x/utils"
)

type emitter[T any] struct {
	kp     *kgo.Client
	serdes mx.Serdes[T]
	out    syncx.Chan[*message_with_ctx]
	done   chan x.Void
}

func (e *emitter[T]) Send(ctx context.Context, message Message[T]) error {
	m, err := e.createMessage(message)
	if err != nil {
		return err
	}

	ctx = utils.IfElse(ctx != nil, ctx, context.Background())
	if !e.out.Write(&message_with_ctx{ctx, m, message.GetAckNack()}) {
		return ErrProducerClosed
	}

	return nil
}

func (e *emitter[T]) Close() {
	e.out.Close()
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

func (e *emitter[T]) createMessage(message Message[T]) (*kgo.Record, error) {
	key, payload, err := e.serdes.Serialize(message.Payload())
	if err != nil {
		return nil, err
	}

	return &kgo.Record{
		Key:   key,
		Value: payload,
		Topic: message.Topic(),
	}, nil
}

func (e *emitter[T]) doSend(mc *message_with_ctx) {
	if mc.ctx.Err() != nil {
		mc.fail(mc.ctx.Err())
		return
	}

	var fn func(record *kgo.Record, err error)
	if mc.ackNack != nil {
		fn = func(record *kgo.Record, err error) {
			mc.completeOrFail(err) // will likely want to send back partition and offset
		}
	}

	e.kp.Produce(mc.ctx, mc.msg, fn)
}

func (e *emitter[T]) begin_processing() {
	defer e.doCleanup()

	for {
		select {
		case <-e.out.ClosedCh():
			return
		case m, ok := <-e.out.Read():
			if !ok {
				return
			} else {
				e.doSend(m)
			}
		}
	}
}

func (e *emitter[T]) doCleanup() {
	e.Close()

	e.kp.Close()

	el, _ := e.out.Drain(nil)
	for _, m := range el {
		m.fail(ErrProducerClosed)
	}

	close(e.done) // signal that emitter is now closed
}

func start[T any](cfg *mx.ClientConfig[T]) (Emitter[T], error) {
	var out syncx.Chan[*message_with_ctx]
	if cfg.Buffer == 0 {
		out = syncx.NewChan[*message_with_ctx]()
	} else {
		out = syncx.NewChanBuffered[*message_with_ctx](cfg.Buffer)
	}

	kp, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ProducerLinger(cfg.Linger),
		kgo.ClientID(cfg.Client_id),
		kgo.SASL(plain.Auth{User: cfg.User, Pass: cfg.Pwd}.AsMechanism()),
		kgo.Dialer(cfg.Dialer.DialContext),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create emitter: %s", err)
	}

	e := &emitter[T]{
		kp:     kp,
		serdes: cfg.Serdes,
		out:    out,
		done:   make(chan x.Void),
	}

	go e.begin_processing()

	return e, nil
}
