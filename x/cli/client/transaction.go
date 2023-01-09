package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/x/crypto"
	"kwil/x/transactions"
	txTypes "kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

// BuildTransaction creates the correct nonce, fee, and signs a transaction
func (c *client) BuildTransaction(ctx context.Context, payloadType transactions.PayloadType, data any, privateKey *ecdsa.PrivateKey) (*txTypes.Transaction, error) {
	// get address from private key
	address, err := crypto.AddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	// get nonce from address
	account, err := c.GetAccount(ctx, address)
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
	price, err := c.EstimatePrice(ctx, tx)
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
