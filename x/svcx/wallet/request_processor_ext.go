package wallet

import (
	"context"
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

	// TODO:  makes more sense to add batch call
	// to emitter and just send a batch from here

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
	w, s, err := deserialize_request(msg)
	if err != nil {
		return async.FailedAction(err)
	}

	if s != nil {
		return r.handle_spend(s, offset)
	}

	return r.handle_withdrawal(w, offset)
}

func (r *request_processor) handle_withdrawal(request *WithdrawalRequest, offset mx.Offset) async.Action {
	// process request event here
	// ...

	// emit confirmation event
	return r.p.Send(context.Background(), request.AsRawEvent())
}

func (r *request_processor) handle_spend(request *SpendRequest, offset mx.Offset) async.Action {
	// process request event here
	// ...

	// emit confirmation event
	return r.p.Send(context.Background(), request.AsRawEvent())
}

func (r *request_processor) is_stopping() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopping
}
