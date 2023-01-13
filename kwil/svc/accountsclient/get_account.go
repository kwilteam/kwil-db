package accountsclient

import (
	"context"
	"fmt"
	"kwil/x/proto/accountspb"
	"kwil/x/types/accounts"
)

func (c *client) GetAccount(ctx context.Context, address string) (accounts.Account, error) {
	res, err := c.accounts.GetAccount(ctx, &accountspb.GetAccountRequest{
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
