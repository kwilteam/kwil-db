package client

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// NewSignedTx creates a signed transaction with a prepared payload. This will
// set the nonce to signer's latest, build the Transaction, set the Fee, and
// sign the transaction. It may then be broadcast on a kwil network. The
// TxOptions may be set to override the nonce and fee.
//
// WARNING: This is an advanced method, and most applications should use the
// other Client methods to interact with a Kwil network.
func (c *Client) NewSignedTx(ctx context.Context, data transactions.Payload, txOpts *clientType.TxOptions) (*transactions.Transaction, error) {
	return c.newTx(ctx, data, txOpts)
}

// newTx creates a new Transaction signed by the Client's Signer
func (c *Client) newTx(ctx context.Context, data transactions.Payload, txOpts *clientType.TxOptions) (*transactions.Transaction, error) {
	if c.Signer == nil {
		return nil, fmt.Errorf("signer must be set to create a transaction")
	}
	if txOpts == nil {
		txOpts = &clientType.TxOptions{}
	}

	var nonce uint64
	if txOpts.Nonce > 0 {
		nonce = uint64(txOpts.Nonce)
	} else {
		// Get the latest nonce for the account, if it exists.
		acc, err := c.txClient.GetAccount(ctx, c.Signer.Identity(), types.AccountStatusPending)
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
	price := txOpts.Fee
	if price == nil {
		price, err = c.txClient.EstimateCost(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate price: %w", err)
		}
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
