package service

import (
	"context"
	"time"

	"github.com/jpillora/backoff"
)

type subscription interface {
	Unsubscribe()
	Err() <-chan error
	Blocks() <-chan int64
}

// Listen will listen to new blocks on the chain.  Confirmed
// determines whether the caller should receive new blocks or only
// confirmed blocks.
func (c *chainClient) Listen(ctx context.Context, confirmed bool) (<-chan int64, error) {
	retChan := make(chan int64)
	sub, err := c.client.Subscribe(ctx, confirmed)
	if err != nil {
		return nil, err
	}

	go func(context.Context, subscription, <-chan int64, bool) {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			return
		case err := <-sub.Err():
			c.log.Errorf("subscription error: %v", err)
			sub = c.resubscribe(ctx, sub, confirmed)
		case <-time.After(c.timeout):
			c.log.Errorf("subscription timeout")
			sub = c.resubscribe(ctx, sub, confirmed)
		case block := <-sub.Blocks():
			if confirmed {
				block -= c.requiredConfirmations
			}
			retChan <- block
		}
	}(ctx, sub, retChan, confirmed)

	return retChan, nil
}

// resubscribe will resubscribe to the chain.  This is used when
// the subscription has an error or is disconnected.
// It will retry forever until it is successful.
func (c *chainClient) resubscribe(ctx context.Context, oldSub subscription, confirmed bool) subscription {
	// unsubscribe from old subscription and create new channel
	oldSub.Unsubscribe()

	// backoff is used to retry for exponential backoffs
	retrier := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	// keep trying to subscribe
	for {
		// exponential backoff
		time.Sleep(retrier.Duration())
		sub, err := c.client.Subscribe(ctx, confirmed)
		if err != nil {
			continue
		}
		retrier.Reset()
		return sub
	}
}
