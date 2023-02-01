package service

import (
	"context"
	"kwil/pkg/chain/provider/dto"
	"time"

	"github.com/jpillora/backoff"
)

// Listen listens to a provider for new blocks.  It will handle disconnections and duplicated / out of order blocks.
func (c *chainClient) Listen(ctx context.Context, blocks chan<- int64) error {
	internalChan := make(chan dto.Header) // receives blocks to be passed to the consumers channel.
	sub, err := c.provider.SubscribeNewHead(ctx, internalChan)
	if err != nil {
		return err
	}

	// set the latest block
	err = c.setLatestBlock(ctx)
	if err != nil {
		return err
	}

	go func(ctx context.Context, c *chainClient, sub dto.Subscription, internalChan chan dto.Header, clientChan chan<- int64) {
		defer sub.Unsubscribe()
		defer close(internalChan)

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				if err != nil {
					c.log.Errorf("subscription error: %v", err)
					sub = c.resubscribe(ctx, sub, internalChan)
				}
			case <-time.After(c.reconnectInterval):
				c.log.Errorf("subscription timeout")
				sub = c.resubscribe(ctx, sub, internalChan)
			case block := <-internalChan:
				height := block.Height - c.requiredConfirmations

				if height <= c.lastBlock {
					c.log.Warnf("received block %d that is less than or equal to the latest block %d", height, c.lastBlock)
					continue
				}

				if height > c.lastBlock+1 {
					c.log.Warnf("received block %d that is greater than the latest block %d by more than 1", height, c.lastBlock)
					for i := c.lastBlock + 1; i < height; i++ {
						clientChan <- i
					}
				}

				c.lastBlock = height
				clientChan <- height
			}
		}
	}(ctx, c, sub, internalChan, blocks)

	return nil
}

// resubscribe will resubscribe to the chain.  This is used when
// the subscription has an error or is disconnected.
// It will retry forever until it is successful.
func (c *chainClient) resubscribe(ctx context.Context, oldSub dto.Subscription, internalChan chan dto.Header) dto.Subscription {
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
		sub, err := c.provider.SubscribeNewHead(ctx, internalChan)
		if err != nil {
			continue
		}
		retrier.Reset()
		return sub
	}
}
