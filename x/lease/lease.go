package lease

import (
	"context"
	"time"
)

const DefaultLeaseDuration = 30 * time.Second
const DefaultHeartbeatFrequency = 15 * time.Second

type Agent interface {
	Subscribe(ctx context.Context, leaseName string, sub Subscriber) error
}

type Lease interface {
	IsRevoked() bool
	OnRevoked() <-chan struct{}
}

type Subscriber struct {
	OnFatalError func(error)
	OnAcquired   func(Lease)
}
