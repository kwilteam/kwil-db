package wallet

import (
	"context"
	"kwil/x"
	"kwil/x/async"
)

// RequestService
// Enacted via gRpc endpoint (all nodes produce)
type RequestService interface {
	SubmitSpend(ctx context.Context, request *SpendRequest) async.Action
	SubmitWithdrawal(ctx context.Context, request *WithdrawalRequest) async.Action

	Start() error
	Stop() error

	OnStop() <-chan x.Void
}
