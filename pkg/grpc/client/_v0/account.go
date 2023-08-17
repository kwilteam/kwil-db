package client

import (
	"context"
	"fmt"

	accountspb "github.com/kwilteam/kwil-db/api/protobuf/accounts/v0"
	"github.com/kwilteam/kwil-db/pkg/accounts"
)

func (c *Client) GetAccount(ctx context.Context, address string) (*accounts.Account, error) {
	res, err := c.accountClt.GetAccount(ctx, &accountspb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	return &accounts.Account{
		Address: res.Account.Address,
		Nonce:   res.Account.Nonce,
		Balance: res.Account.Balance,
		Spent:   res.Account.Spent,
	}, nil
}
