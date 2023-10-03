package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

// newTx creates a new Transaction signed by the Client's Signer
func (c *Client) newTx(ctx context.Context, data transactions.Payload) (*transactions.Transaction, error) {
	// get nonce from address
	acc, err := c.transportClient.GetAccount(ctx, c.Signer.PublicKey())
	if err != nil {
		acc = &balances.Account{
			PublicKey: c.Signer.PublicKey(),
			Nonce:     0,
			Balance:   big.NewInt(0),
		}
	}

	// build transaction
	tx, err := transactions.CreateTransaction(data, uint64(acc.Nonce+1))
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// estimate price
	price, err := c.transportClient.EstimateCost(ctx, tx)
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
