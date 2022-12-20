package client

import (
	"context"
	"fmt"
	"kwil/x/chain-client/dto"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// Subscribe subscribes to new blocks and returns a subscription
func (e *EVMClient) Subscribe(ctx context.Context) (dto.Subscription, error) {
	headerChan := make(chan *types.Header) // this will be converted to a block height channel
	sub, err := e.client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new block headers: %v", err)
	}

	// create the subscription
	subscription := &EVMSubscription{
		blocks: make(chan int64),
		errs:   make(chan error),
		sub:    sub,
	}

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
