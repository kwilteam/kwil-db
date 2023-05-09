package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
)

// buildTx creates the correct nonce, fee, and signs a transaction
func (c *KwilClient) buildTx(ctx context.Context, payloadType kTx.PayloadType, data any, privateKey *ecdsa.PrivateKey) (*kTx.Transaction, error) {
	// serialize data
	bts, err := serialize.Serialize(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	address, err := crypto.AddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	// get nonce from address
	acc, err := c.grpc.GetAccount(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account config: %w", err)
	}

	// build transaction
	tx := kTx.NewTx(payloadType, bts, acc.Nonce+1)

	// estimate price
	price, err := c.grpc.EstimateCost(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	// set fee
	tx.Fee = price.String()

	// sign transaction
	err = tx.Sign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}
