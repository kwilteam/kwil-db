package client

import (
	"context"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types"
)

func (c *Client) GetAccount(ctx context.Context, pubKey []byte) (*types.Account, error) {
	res, err := c.txClient.GetAccount(ctx, &txpb.GetAccountRequest{
		PublicKey: pubKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	bigBal, ok := new(big.Int).SetString(res.Account.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse balance")
	}

	acc := &types.Account{
		PublicKey: res.Account.PublicKey,
		Balance:   bigBal,
		Nonce:     res.Account.Nonce,
	}

	return acc, nil
}
