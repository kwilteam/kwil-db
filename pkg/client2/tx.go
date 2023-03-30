package client2

import (
	"context"
	"fmt"
	"kwil/pkg/crypto"
	kTx "kwil/pkg/tx"
	"kwil/pkg/utils/serialize"
)

func (c *Client) newTx(ctx context.Context, payloadType kTx.PayloadType, data any) (*kTx.Transaction, error) {
	if c.PrivateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	// serialize data
	bts, err := serialize.Serialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	// get nonce from address
	acc, err := c.client.GetAccount(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account config: %w", err)
	}

	// build transaction
	tx := kTx.NewTx(payloadType, bts, acc.Nonce+1)

	// estimate price
	price, err := c.client.EstimateCost(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// set fee
	tx.Fee = price.String()

	// sign transaction
	err = tx.Sign(c.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

func (c *Client) getAddress() (string, error) {
	if c.PrivateKey == nil {
		return "", fmt.Errorf("private key is nil")
	}

	return crypto.AddressFromPrivateKey(c.PrivateKey)
}
