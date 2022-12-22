package service

import (
	"context"
	"time"

	provider "kwil/x/chain/provider/dto"

	"github.com/jpillora/backoff"
)

// Listen listens to a provider for new blocks.  It will handle disconnections and duplicated / out of order blocks.
func (c *chainClient) Listen(ctx context.Context, blocks chan<- provider.Header) error {
	internalChan := make(chan provider.Header) // receives blocks to be passed to the consumers channel.
	sub, err := c.provider.SubscribeNewHead(ctx, internalChan)
	if err != nil {
		return err
	}

	go func(ctx context.Context, sub provider.Subscription, internalChan chan provider.Header, clientChan chan<- provider.Header) {
		defer sub.Unsubscribe()
		defer close(internalChan)

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				if err != nil {
					c.log.Errorf("subscription error: %v", err)
					sub = c.resubscribe(ctx, sub, clientChan)
				}
			case <-time.After(c.maxBlockInterval):
				c.log.Errorf("subscription timeout")
				sub = c.resubscribe(ctx, sub, clientChan)
			case block := <-internalChan:
				clientChan <- block
			}
		}
	}(ctx, sub, internalChan, blocks)

	return nil
}

// resubscribe will resubscribe to the chain.  This is used when
// the subscription has an error or is disconnected.
// It will retry forever until it is successful.
func (c *chainClient) resubscribe(ctx context.Context, oldSub provider.Subscription, clientChan chan<- provider.Header) provider.Subscription {
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
		sub, err := c.provider.SubscribeNewHead(ctx, clientChan)
		if err != nil {
			continue
		}
		retrier.Reset()
		return sub
	}
}
