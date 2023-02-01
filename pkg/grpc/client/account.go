package client

import (
	"context"
	"fmt"
	pb "kwil/api/protobuf/account/v0/gen/go"
	"kwil/x/types/accounts"
)

func (c *Client) GetAccount(ctx context.Context, address string) (accounts.Account, error) {
	res, err := c.accountClt.GetAccount(ctx, &pb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return accounts.Account{}, fmt.Errorf("failed to get account: %w", err)
	}

	return accounts.Account{
		Address: res.Account.Address,
		Nonce:   res.Account.Nonce,
		Balance: res.Account.Balance,
		Spent:   res.Account.Spent,
	}, nil
}
