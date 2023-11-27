package client

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// newTx creates a new Transaction signed by the Client's Signer
func (c *Client) newTx(ctx context.Context, data transactions.Payload, opts ...TxOpt) (*transactions.Transaction, error) {
	txOpts := &txOptions{}
	for _, opt := range opts {
		opt(txOpts)
	}

	var nonce uint64
	if txOpts.nonce > 0 {
		nonce = uint64(txOpts.nonce)
	} else {
		// Get the latest nonce for the account, if it exists.
		acc, err := c.rpc.GetAccount(ctx, c.Signer.Identity(), types.AccountStatusPending)
		if err != nil {
			return nil, err
		}
		// NOTE: an error type would be more robust signalling of a non-existent
		// account, but presently a nil ID is set by internal/accounts.
		if len(acc.Identifier) > 0 {
			nonce = uint64(acc.Nonce + 1)
		} else {
			nonce = 1
		}
	}

	// build transaction
	tx, err := transactions.CreateTransaction(data, c.chainID, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// estimate price
	price, err := c.rpc.EstimateCost(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// set fee
	tx.Body.Fee = price

	// sign transaction
	err = tx.Sign(c.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}