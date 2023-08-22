package client

import (
	"context"
	"fmt"
	"math/big"

	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

// Transaction signed by the client
func (c *Client) newTx(ctx context.Context, data transactions.Payload) (*transactions.Transaction, error) {
	if c.Signer == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	// get nonce from address
	acc, err := c.client.GetAccount(ctx, address)
	if err != nil {
		acc = &balances.Account{
			Address: address,
			Nonce:   0,
			Balance: big.NewInt(0),
		}
	}

	// build transaction
	tx, err := transactions.CreateTransaction(data, uint64(acc.Nonce+1))
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// estimate price
	price, err := c.client.EstimateCost(ctx, tx)
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

// Tx Signed by the Validator Node
//
// In the transaction, both the signature and sender are set differently than if
// the usual Sign method were used:
//   - Signature is an Ed25519 signature
//   - Sender is the base64-encoded *pubkey*, not an address.
func (c *Client) NewNodeTx(ctx context.Context, payload transactions.Payload, privKey []byte) (*transactions.Transaction, error) {

	nodeKey := cmtEd.PrivKey(privKey)

	pubKey := nodeKey.PubKey()
	address := pubKey.Address().String()
	acc, err := c.client.GetAccount(ctx, address)
	if err != nil {
		acc = &balances.Account{
			Address: address,
			Nonce:   0,
			Balance: big.NewInt(0),
		}
	}

	// build transaction
	tx, err := transactions.CreateTransaction(payload, uint64(acc.Nonce+1))
	if err != nil {
		return nil, err
	}

	// sign transaction
	hash, err := tx.SetHash()
	if err != nil {
		return nil, err
	}

	// Sender is not in the hash. Move this if that changes.
	tx.Sender = pubKey.Bytes()

	sign, err := nodeKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	tx.Signature = &crypto.Signature{
		Signature: sign,
		Type:      crypto.SignatureTypeEd25519,
	}

	fmt.Println("tx hash", hash)
	fmt.Println("tx sender", tx.Sender)
	fmt.Println("tx signature", tx.Signature)
	fmt.Println("tx payload", tx.Body.Payload)
	fmt.Println("tx payload type", tx.Body.PayloadType)

	return tx, nil
}

func (c *Client) getAddress() (string, error) {
	if c.Signer == nil {
		return "", fmt.Errorf("private key is nil")
	}

	return c.Signer.PubKey().Address().String(), nil
}
