package wallet

import (
	"kwil/x"
	"kwil/x/async"
	"kwil/x/svcx/messaging/mx"
)

// RequestEvents
// Consumed by processor
type RequestEvents interface {
	OnWithdrawal(func(WithdrawalEvent) async.Action)
	OnSpend(func(SpendEvent) async.Action)

	Start() error
	Stop() error

	OnStop() <-chan x.Void
}

type WithdrawalEvent struct {
	request_id string
	// ...
}

type SpendEvent struct {
	request_id string
	// ...
}

func (s *SpendEvent) AsMessage() *mx.RawMessage {
	// wallet id as key
	// request as value (need to include type as a marker in order to deserialize later during processing)
	panic("implement me")
}

func (s *WithdrawalEvent) AsMessage() *mx.RawMessage {
	// wallet id as key
	// request as value (need to include type as a marker in order to deserialize later during processing)
	panic("implement me")
}
