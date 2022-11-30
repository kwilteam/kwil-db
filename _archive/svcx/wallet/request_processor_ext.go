package wallet

import (
	"context"
	"fmt"
	"kwil/archive/svcx/messaging/mx"
	"kwil/archive/svcx/messaging/pub"
	"kwil/archive/svcx/messaging/sub"
	"kwil/x"
	"kwil/x/async"
	"sync"
)

type request_processor struct {
	p         pub.ByteEmitter
	e         sub.TransientReceiver
	done      chan x.Void
	stop      chan x.Void
	transform MessageTransform
	wg        *sync.WaitGroup
	mu        *sync.Mutex
	stopping  bool
}

func (r *request_processor) Start() error {
	err := r.e.Start()
	if err != nil {
		return err
	}

	go r.run()
	return nil
}

func (r *request_processor) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopping {
		return fmt.Errorf("already stopping")
	}

	r.stopping = true

	close(r.stop)

	return nil
}

func (r *request_processor) OnStop() <-chan x.Void {
	return r.done
}

func (r *request_processor) run() {
	for !r.is_stopping() {
		select {
		case <-r.stop:
			return
		case it := <-r.e.OnReceive():
			if r.is_stopping() {
				it.Commit()
				return
			}
			r.wg.Add(1)
			go r.handle_messages(it)
		}
	}

	r.wg.Wait()

	r.p.Close()
	r.e.Stop()

	<-r.e.OnStop()

	close(r.done)
}

func (r *request_processor) handle_messages(iter sub.MessageIterator) {
	if !iter.HasNext() {
		iter.Commit().WhenComplete(r.handle_if_error_and_set_wg_done)
		return
	}

	msg, offset := iter.Next()

	r.handle_message(msg, offset).
		OnCompleteA(&async.ContinuationA{
			Then:  r.get_next(iter),
			Catch: r.handle_if_error_and_set_wg_done,
		})
}

func (r *request_processor) handle_message(msg *mx.RawMessage, offset mx.Offset) async.Action {
	msg, request_id, err := decode_message(msg)
	if err != nil {
		return async.FailedAction(err)
	}

	return r.handle(msg, offset, request_id)
}

func (r *request_processor) handle(msg *mx.RawMessage, _ mx.Offset, request_id string) async.Action {
	return r.transform(msg).ComposeA(r.on_transform(request_id))
}

func (r *request_processor) on_transform(request_id string) func(msg *mx.RawMessage, err error) async.Action {
	return func(msg *mx.RawMessage, err error) async.Action {
		if err != nil {
			return async.FailedAction(err)
		}

		if request_id == "" {
			return async.CompletedAction()
		}

		// emit confirmation event
		return r.p.Send(context.Background(), encode_event(request_id, msg))
	}
}

func (r *request_processor) is_stopping() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopping
}

func (r *request_processor) get_next(iter sub.MessageIterator) func() {
	return func() {
		// iterate next item once item has been emitted
		r.handle_messages(iter)
	}
}

func (r *request_processor) handle_if_error_and_set_wg_done(err error) {
	defer r.wg.Done()
	if err == nil {
		return
	}

	fmt.Printf("error handling message: %v\n", err)
	_ = r.Stop()
}
