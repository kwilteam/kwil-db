package wallet

import (
	"context"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/svcx/messaging/pub"
)

type request_processor struct {
	p pub.ByteEmitter
	e RequestEvents
}

func (r *request_processor) Start() error {
	err := r.e.Start()
	if err != nil {
		return err
	}

	r.e.OnWithdrawal(r.onWithdrawal)
	r.e.OnSpend(r.onSpend)

	return nil
}

func (r *request_processor) Stop() error {
	r.p.Close()
	return r.e.Stop()
}

func (r *request_processor) OnStop() <-chan x.Void {
	return r.e.OnStop()
}

func (r *request_processor) onWithdrawal(event WithdrawalEvent) async.Action {
	// process request event here
	// ...

	// emit confirmation event
	return r.p.Send(context.Background(), event.AsMessage())
}

func (r *request_processor) onSpend(event SpendEvent) async.Action {
	// process request event here
	// ...

	// emit confirmation event
	return r.p.Send(context.Background(), event.AsMessage())
}
