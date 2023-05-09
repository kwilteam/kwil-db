package client

import (
	"context"
	"github.com/kwilteam/kwil-db/pkg/balances"
)

func (c *KwilClient) GetAccount(ctx context.Context, address string) (*balances.Account, error) {
	return c.grpc.GetAccount(ctx, address)
}
