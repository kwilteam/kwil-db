package grpc_client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/x/crypto"
	txTypes "kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

// BuildTransaction creates the correct nonce, fee, and signs a transaction
func (c *Client) BuildTransaction(ctx context.Context, payloadType txTypes.PayloadType, data any, privateKey *ecdsa.PrivateKey) (*txTypes.Transaction, error) {
	address, err := crypto.AddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	// get nonce from address
	account, err := c.Accounts.GetAccount(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// serialize data
	bts, err := serialize.Serialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	// build transaction
	tx := txTypes.NewTx(payloadType, bts, account.Nonce+1)

	// estimate price
	price, err := c.Pricing.EstimateCost(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// set fee
	tx.Fee = price

	// sign transaction
	err = tx.Sign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}
