package wallet

import (
	"context"
	"fmt"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/svcx/messaging/mx"
	"kwil/x/svcx/messaging/pub"
	"kwil/x/svcx/messaging/sub"
	"sync"
)

type request_processor struct {
	p        pub.ByteEmitter
	e        sub.TransientReceiver
	done     chan x.Void
	stop     chan x.Void
	wg       *sync.WaitGroup
	mu       *sync.Mutex
	stopping bool
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
		return nil
	}

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
		iter.Commit().WhenComplete(r.on_iter_complete)
		return
	}

	// TODO: makes more sense to add batch call
	// to emitter and just send a batch from here

	msg, offset := iter.Next()
	r.handle_message(msg, offset).
		OnCompleteA(&async.ContinuationA{
			Then:  r.get_next(iter),
			Catch: r.handle_if_error,
		})
}

func (r *request_processor) handle_message(msg *mx.RawMessage, offset mx.Offset) async.Action {
	msg, request_id, err := decode_message(msg)
	if err != nil {
		return async.FailedAction(err)
	}

	return r.handle(msg, offset, request_id)
}

func (r *request_processor) handle(msg *mx.RawMessage, offset mx.Offset, request_id string) async.Action {
	// process request event here
	// ...

	if request_id == "" {
		return async.CompletedAction()
	}

	// emit confirmation event
	return r.p.Send(context.Background(), encode_event(request_id, msg))
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

func (r *request_processor) on_iter_complete(err error) {
	r.wg.Done()
	r.handle_if_error(err)
}

func (r *request_processor) handle_if_error(err error) {
	if err == nil {
		return
	}

	fmt.Printf("error handling message: %v\n", err)
	r.wg.Done()
	_ = r.Stop()
}
