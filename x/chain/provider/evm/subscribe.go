package evm

import (
	"context"
	"fmt"
	"kwil/x/chain/provider/dto"

	"github.com/ethereum/go-ethereum/core/types"
)

func (c *ethClient) SubscribeNewHead(ctx context.Context, ch chan<- dto.Header) (dto.Subscription, error) {

	ethHeaderChan := make(chan *types.Header)

	sub, err := c.ethclient.SubscribeNewHead(ctx, ethHeaderChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new EVM block headers: %v", err)
	}

	newSub := newEthSubscription(sub)

	// we will listen and convert the headers to our own headers/
	// this is simply a passthrough
	go func(ctx context.Context, ethHeaderChan <-chan *types.Header, ch chan<- dto.Header, sub *ethSubscription) {
		for {
			select {
			case ethHeader := <-ethHeaderChan:
				ch <- dto.Header{
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
