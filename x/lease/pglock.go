package lease

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	plck "cirello.io/pglock"
	"github.com/google/uuid"
)

type agent struct {
	c *plck.Client
}

func NewAgent(db *sql.DB, owner string) (Agent, error) {
	ower_suffix := uuid.New().String()
	c, err := plck.UnsafeNew(db,
		plck.WithLeaseDuration(DefaultLeaseDuration),
		plck.WithHeartbeatFrequency(DefaultHeartbeatFrequency),
		plck.WithCustomTable("distributed_locks"),
		plck.WithOwner(owner+"_"+ower_suffix),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create pglock agent: %w", err)
	}

	return &agent{c: c}, nil
}

func (a *agent) Subscribe(ctx context.Context, leaseName string, sub Subscriber) error {
	if leaseName == "" {
		return fmt.Errorf("lease name cannot be empty")
	}

	go func() {
		for {
			err := a.c.Do(ctx, leaseName, func(inner_ctx context.Context, lock *plck.Lock) error {
				sub.OnAcquired(&pg_lease{
					f: lock.IsReleased,
					r: inner_ctx.Done(),
				})
				return inner_ctx.Err() // if cancelled, then the lease was revoked
			})

			// if nil, the lease is no longer needed
			if err == nil {
				return
			}

			// cancelled by caller, so exit
			if ctx.Err() != nil {
				return
			}

			if !errors.Is(err, context.Canceled) {
				sub.OnFatalError(err) // fatal error if not a cancellation
				return
			}
		}
	}()

	return nil
}

type pg_lease struct {
	f func() bool
	r <-chan struct{}
}

func (l *pg_lease) IsRevoked() bool {
	return l.f()
}

func (l *pg_lease) OnRevoked() <-chan struct{} {
	return l.r
}
