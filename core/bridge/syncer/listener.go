package syncer

import (
	"context"
	"fmt"
	"time"

	"github.com/jpillora/backoff"
	cClient "github.com/kwilteam/kwil-db/core/chain"
	"github.com/kwilteam/kwil-db/core/types/chain"
)

func (b *blockSyncer) Listen(ctx context.Context, blocks chan<- int64) error {
	headChan := make(chan chain.Header)
	sub, err := b.chainClient.SubscribeNewHead(ctx, headChan)
	if err != nil {
		return err
	}

	// set the latest block
	err = b.setLatestBlock(ctx)
	if err != nil {
		return err
	}

	go func(ctx context.Context, bc cClient.ChainClient, sub chain.Subscription, headChan chan chain.Header, blockChan chan<- int64) {
		defer sub.Unsubscribe()
		defer close(headChan)

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				if err != nil {
					fmt.Println("subscription error", "err", err)
					sub = b.resubscribe(ctx, sub, headChan)
				}
			case <-time.After(b.reconnectInterval):
				fmt.Println("subscription timeout")
				sub = b.resubscribe(ctx, sub, headChan)

			case block := <-headChan:
				height := block.Height - b.requiredConfirmations

				if height <= b.lastBlock {
					continue
				}

				if height > b.lastBlock+1 {
					for i := b.lastBlock + 1; i < height; i++ {
						blockChan <- i
					}
				}

				b.lastBlock = height
				blockChan <- height
			}
		}
	}(ctx, b.chainClient, sub, headChan, blocks)

	return nil
}

// resubscribe will resubscribe to the chain.  This is used when
// the subscription has an error or is disconnected.
// It will retry forever until it is successful.
func (b *blockSyncer) resubscribe(ctx context.Context, oldSub chain.Subscription, internalChan chan chain.Header) chain.Subscription {
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
		sub, err := b.chainClient.SubscribeNewHead(ctx, internalChan)
		if err != nil {
			continue
		}
		retrier.Reset()
		return sub
	}
}
