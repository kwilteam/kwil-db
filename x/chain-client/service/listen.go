package service

import (
	"context"
	"kwil/x/chain-client/dto"
	"time"

	"github.com/jpillora/backoff"
)

func (c *chainClient) Listen(ctx context.Context, confirmed bool) (<-chan int64, error) {
	retChan := make(chan int64)
	sub, err := c.listener.Subscribe(ctx)
	if err != nil {
		return nil, err
	}

	go func(ctx context.Context, sub dto.Subscription, retChan chan int64, confirmed bool) {
		defer sub.Unsubscribe()
		defer close(retChan)

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				if err != nil {
					c.log.Errorf("subscription error: %v", err)
					sub = c.resubscribe(ctx, sub, confirmed)
				}
			case <-time.After(c.maxBlockInterval):
				c.log.Errorf("subscription timeout")
				sub = c.resubscribe(ctx, sub, confirmed)
			case block := <-sub.Blocks():
				if confirmed {
					block -= c.requiredConfirmations
				}
				retChan <- block
			}
		}
	}(ctx, sub, retChan, confirmed)

	return retChan, nil
}

// resubscribe will resubscribe to the chain.  This is used when
// the subscription has an error or is disconnected.
// It will retry forever until it is successful.
func (c *chainClient) resubscribe(ctx context.Context, oldSub dto.Subscription, confirmed bool) dto.Subscription {
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
		sub, err := c.listener.Subscribe(ctx)
		if err != nil {
			continue
		}
		retrier.Reset()
		return sub
	}
}
