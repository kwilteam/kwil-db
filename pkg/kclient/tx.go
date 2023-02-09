package kclient

import (
	"context"
	"fmt"
	"kwil/pkg/accounts"
	"kwil/pkg/utils/serialize"
)

// buildTx creates the correct nonce, fee, and signs a transaction
func (c *Client) buildTx(ctx context.Context, account string, payloadType accounts.PayloadType, data any) (*accounts.Transaction, error) {
	// serialize data
	bts, err := serialize.Serialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	// get nonce from address
	acc, err := c.Kwil.GetAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to get account config: %w", err)
	}

	// build transaction
	tx := accounts.NewTx(payloadType, bts, acc.Nonce+1)

	// estimate price
	price, err := c.Kwil.EstimateCost(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// set fee
	tx.Fee = price

	// sign transaction
	err = tx.Sign(c.Config.Fund.Wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}
