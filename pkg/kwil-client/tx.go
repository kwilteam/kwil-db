package kwil_client

import (
	"context"
	"fmt"
	transactions2 "kwil/pkg/types/transactions"
	"kwil/pkg/utils/serialize"
)

// buildTx creates the correct nonce, fee, and signs a transaction
func (c *Client) buildTx(ctx context.Context, account string, payloadType transactions2.PayloadType, data any) (*transactions2.Transaction, error) {
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
	tx := transactions2.NewTx(payloadType, bts, acc.Nonce+1)

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
