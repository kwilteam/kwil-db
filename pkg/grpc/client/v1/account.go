package client

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/pkg/balances"
	"kwil/pkg/utils/serialize"
)

func (c *Client) GetAccount(ctx context.Context, address string) (*balances.Account, error) {
	res, err := c.txClient.GetAccount(ctx, &txpb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	acc, err := serialize.Convert[txpb.Account, balances.Account](res.Account)
	if err != nil {
		return nil, fmt.Errorf("failed to convert account: %w", err)
	}

	return acc, nil
}
