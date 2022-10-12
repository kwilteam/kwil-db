package pub

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"kwil/x"
	"kwil/x/messaging/internal"
	"kwil/x/messaging/mx"
	"kwil/x/syncx"
	"kwil/x/utils"
	"sync"
)

// emitter_client the client is decoupled from the
// emitter since the client itself can multiplex
// to the same cluster with multiple emitters.
type emitter_client struct {
	kp       *kgo.Client
	out      syncx.Chan[*message_with_ctx]
	done     syncx.Chan[x.Void]
	mu       sync.Mutex
	emitters map[int]internal.Closable
}

func (e *emitter_client) GetClientType() mx.ClientType {
	return mx.Emitter
}

func (e *emitter_client) IsClosed() bool {
	return e.done.IsClosed()
}

func (e *emitter_client) Close() bool {
	return e.out.Close()
}

func (e *emitter_client) OnClosed() <-chan x.Void {
	return e.done.ClosedCh()
}

func (e *emitter_client) closeAndWait(ctx context.Context) error {
	e.Close()
	select {
	case <-e.done.ClosedCh():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *emitter_client) send_async(ctx context.Context, message *kgo.Record, ack_nack AckNackFn) error {
	var a *AckNackFn
	if ack_nack != nil {
		a = &ack_nack
	}

	ctx = utils.IfElse(ctx != nil, ctx, context.Background())
	if !e.out.Write(&message_with_ctx{ctx, message, a}) {
		return ErrProducerClosed
	}

	return nil
}

func (e *emitter_client) send(mc *message_with_ctx) {
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

func (e *emitter_client) begin_processing() {
	defer e.doCleanup()

	for {
		select {
		case <-e.out.ClosedCh():
			return
		case m, ok := <-e.out.Read():
			if !ok {
				return
			} else {
				e.send(m)
			}
		}
	}
}

func (e *emitter_client) doCleanup() {
	e.Close()

	e.kp.Close()

	el, _ := e.out.Drain(nil)
	for _, m := range el {
		// not sure if we should fail here or not
		// closing the client itself below may be
		// sufficient and make more sense to handle
		// within a more specific context.
		m.fail(ErrProducerClosed)
	}

	e.mu.Lock()
	emitters := e.emitters
	e.emitters = make(map[int]internal.Closable)
	e.mu.Unlock()

	// close and wait for each underlying emitter
	// to close out.
	for id, e := range emitters {
		delete(emitters, id)
		e.Close()
		<-e.OnClosed()
	}

	// signal that client is now closed
	e.done.Close()
}

func (e *emitter_client) attach(emitter internal.Closable) (func() <-chan x.Void, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	current := e.emitters[emitter.ID()]
	if current != nil {
		return nil, fmt.Errorf("emitter with id %d has already been associated with this client", emitter.ID())
	}

	e.emitters[emitter.ID()] = emitter

	// Capturing each variable here to remove direct
	// references to the client. This is to prevent
	// the client from being referenced in the map
	// stored function in case the emitter is not
	// properly closed. Maybe being paranoid here.
	id := emitter.ID()
	out := e.out
	mu := &e.mu
	done := e.done
	emitters := e.emitters

	return func() <-chan x.Void {
		mu.Lock()
		defer mu.Unlock()
		inner := emitters[id]
		if inner == nil {
			// in case of a cyclic shutdown (e.g., emitter.Close() ->
			// client.Close() -> emitter.Close(), etc.) we check for
			// a nil emitter here.
			return done.ClosedCh()
		}

		delete(emitters, id)
		if len(emitters) > 0 {
			return x.ClosedChanVoid()
		}

		out.Close()

		return done.ClosedCh()
	}, nil
}
