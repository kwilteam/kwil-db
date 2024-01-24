package client

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

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
	price, err := c.txClient.EstimateCost(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// c.logger.Info("cost estimate", zap.String("fee", price.String()), zap.Uint64("nonce", nonce))

	// set fee
	tx.Body.Fee = price

	// sign transaction
	err = tx.Sign(c.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// b, _ := tx.MarshalBinary()
	// c.logger.Info(fmt.Sprintf("tx hash %x", sha256.Sum256(b)))

	return tx, nil
}
