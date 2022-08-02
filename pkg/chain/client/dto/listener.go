package dto

import "context"

type Listener interface {
	Subscribe(ctx context.Context) (Subscription, error)
	GetLatestBlock(ctx context.Context) (int64, error)
}
