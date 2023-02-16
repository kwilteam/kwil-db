package client

import (
	"context"
	"kwil/pkg/accounts"
)

func (c *KwilClient) GetAccount(ctx context.Context, address string) (*accounts.Account, error) {
	return c.grpc.GetAccount(ctx, address)
}
