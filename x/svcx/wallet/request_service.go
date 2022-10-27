package wallet

import (
	"context"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/svcx/messaging/mx"
)

// RequestService
// Enacted via gRpc endpoint (all nodes produce)
type RequestService interface {
	Submit(ctx context.Context, msg *mx.RawMessage) async.Action
	SubmitAsync(ctx context.Context, msg *mx.RawMessage) async.Action

	Start() error
	Stop() error

	OnStop() <-chan x.Void
}
