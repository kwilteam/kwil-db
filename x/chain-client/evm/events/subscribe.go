package events

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

func (e *EVMEventListener) Subscribe(ctx context.Context) (*EVMSubscription, error) {
	headerChan := make(chan *types.Header) // this will be converted to a block height channel
	sub, err := e.client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new block headers: %v", err)
	}

	// create the subscription
	subscription := newSub(sub)

	// goroutine to receive block heights and return them to the subscription
	go func(context.Context, *EVMSubscription, ethereum.Subscription, chan *types.Header) {
		for {
			// we will listen to new blocks, errors, and context cancellation
			// we don't need to reconnect here; this happens in the service
			select {
			case header := <-headerChan:
				subscription.blocks <- header.Number.Int64()
			case <-ctx.Done():
				return
			case err := <-sub.Err(): // this should return when we unsubscribe from the ethclient (see unsubscribe.go)
				subscription.errs <- err
			}
		}
	}(ctx, subscription, sub, headerChan)

	return subscription, nil
}
