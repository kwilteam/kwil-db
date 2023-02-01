package kcli

import (
	"context"
	"fmt"
	"kwil/x/types/accounts"
	"kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

// buildTx creates the correct nonce, fee, and signs a transaction
func (c *KwilClient) buildTx(ctx context.Context, account accounts.Account, payloadType transactions.PayloadType, data any) (*transactions.Transaction, error) {
	// serialize data
	bts, err := serialize.Serialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	// build transaction
	tx := transactions.NewTx(payloadType, bts, account.Nonce+1)

	// estimate price
	price, err := c.Client.EstimateCost(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// set fee
	tx.Fee = price

	// sign transaction
	err = tx.Sign(c.cfg.Fund.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}
