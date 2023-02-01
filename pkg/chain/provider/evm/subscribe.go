package evm

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	dto2 "kwil/pkg/chain/provider/dto"
)

func (c *ethClient) SubscribeNewHead(ctx context.Context, ch chan<- dto2.Header) (dto2.Subscription, error) {

	ethHeaderChan := make(chan *types.Header)

	sub, err := c.ethclient.SubscribeNewHead(ctx, ethHeaderChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new EVM block headers: %v", err)
	}

	newSub := newEthSubscription(sub)

	// we will listen and convert the headers to our own headers/
	// this is simply a passthrough
	go func(ctx context.Context, ethHeaderChan <-chan *types.Header, ch chan<- dto2.Header, sub *ethSubscription) {
		for {
			select {
			case ethHeader := <-ethHeaderChan:
				ch <- dto2.Header{
					Height: ethHeader.Number.Int64(),
					Hash:   ethHeader.Hash().Hex(),
				}
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				newSub.errs <- err
			}
		}
	}(ctx, ethHeaderChan, ch, newSub)

	return newSub, nil
}
