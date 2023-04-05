package client

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/pkg/balances"
	"math/big"
)

func (c *Client) GetAccount(ctx context.Context, address string) (*balances.Account, error) {
	res, err := c.txClient.GetAccount(ctx, &txpb.GetAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	bigBal, ok := new(big.Int).SetString(res.Account.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance")
	}

	acc := &balances.Account{
		Address: res.Account.Address,
		Balance: bigBal,
		Nonce:   res.Account.Nonce,
	}

	return acc, nil
}
