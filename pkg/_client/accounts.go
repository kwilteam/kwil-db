package client

import (
	"context"
	"kwil/pkg/balances"
)

func (c *KwilClient) GetAccount(ctx context.Context, address string) (*balances.Account, error) {
	return c.grpc.GetAccount(ctx, address)
}
