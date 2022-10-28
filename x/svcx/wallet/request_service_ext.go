package wallet

import (
	"context"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/svcx/messaging/mx"
	"kwil/x/svcx/messaging/pub"
	"sync"
)

// implements RequestService
type request_Service struct {
	p         pub.ByteEmitter
	c         ConfirmationEvents
	mu        sync.Mutex
	responses map[string]async.Action // TODO: will need to add map clean-up for response failures
}

func (r *request_Service) SubmitAsync(ctx context.Context, msg *mx.RawMessage) async.Action {
	return r.p.Send(ctx, encode_message_async(msg))
}
func (r *request_Service) Submit(ctx context.Context, msg *mx.RawMessage) async.Action {
	response := async.NewAction()

	msg, request_id := encode_message(msg)
	r.p.
		Send(ctx, msg). // send to topic
		OnCompleteA(&async.ContinuationA{
			Then: func() {
				// broker has acknowledged receipt of message
				// store message and await confirmation event
				r.add_response(request_id, response)
			},
			Catch: func(err error) {
				// unexpected failure sending message
				// underlying client has retry logic
				// for transient errors, so this could
				// be a permanent failure or a timeout
				// NOTE: we will likely want to have the
				// background service interrogate the response
				// more fully to continue long-running retries
				// due to a major outage (e.g., no need to
				// stop if the broker wil eventually come back up
				// alternatively, kubernetes could be configured
				// to orchestrate this behavior -- which is the
				// better practice)
				response.Fail(err)
			},
		})

	return response
}

func (r *request_Service) Start() error {
	err := r.c.Start()
	if err != nil {
		return err
	}

	return nil
}

func (r *request_Service) Stop() error {
	r.p.Close()
	return r.c.Stop()
}

func (r *request_Service) OnStop() <-chan x.Void {
	return r.c.OnStop()
}

func (r *request_Service) handle_event_response(ev ConfirmationEvent) async.Action {
	return r.signal_response(ev.request_id)
}

func (r *request_Service) add_response(request_id string, response async.Action) {
	r.mu.Lock()
	r.responses[request_id] = response
	r.mu.Unlock()
}

func (r *request_Service) signal_response(request_id string) async.Action {
	r.mu.Lock()
	response := r.responses[request_id]
	if response == nil {
		return async.CompletedAction()
	}

	delete(r.responses, request_id)
	r.mu.Unlock()

	response.Complete()

	return async.CompletedAction()
}
