package wallet

import (
	"context"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/deposits/processor"
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
	pr       processor.Processor
}

func (r *request_processor) Start() error {
	err := r.e.Start()
	if err != nil {
		return err
	}

	err = r.e.Start()
	if err != nil {
		r.e.Stop()
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
		r.wg.Done()
	}

	msg, offset := iter.Next()

	r.handle_message(msg, offset).
		OnCompleteA(&async.ContinuationA{
			Then: func() {
				// iterate next item once item has been emitted
				r.handle_messages(iter)
			},
			Catch: func(err error) {
				// log error
				r.wg.Done()
				_ = r.Stop()
			},
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
