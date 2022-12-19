package client

import (
	"context"
	"kwil/x/chain-client/evm/events"
)

func (c *EVMClient) Subscribe(ctx context.Context) (*events.EVMSubscription, error) {
	return c.listener.Subscribe(ctx)
}
