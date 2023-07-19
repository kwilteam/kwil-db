package client

import (
	"context"
	"encoding/json"
	"fmt"

	cmtCrypto "github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

// Transaction signed by the client
func (c *Client) newTx(ctx context.Context, payloadType kTx.PayloadType, data any) (*kTx.Transaction, error) {
	if c.PrivateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	// serialize data
	bts, err := json.Marshal(data)
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

// Tx Signed by the Validator Node
func (c *Client) NewNodeTx(ctx context.Context, payloadType kTx.PayloadType, data any, privKey string) (*kTx.Transaction, error) {
	var nodeKey cmtCrypto.PrivKey
	key := fmt.Sprintf(`{"type":"tendermint/PrivKeyEd25519","value":"%s"}`, privKey)
	err := cmtjson.Unmarshal([]byte(key), &nodeKey)
	if err != nil {
		return nil, err
	}

	// serialize data
	bts, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data: %w", err)
	}

	// build transaction
	tx := kTx.NewTx(payloadType, bts, 1)
	// sign transaction
	hash := tx.GenerateHash()
	sign, err := nodeKey.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	var keytype crypto.SignatureType
	if nodeKey.Type() == "ed25519" {
		keytype = crypto.PK_ED25519
	}

	tx.Signature = &crypto.Signature{
		Signature: sign,
		Type:      keytype,
	}

	tx.Hash = hash
	keybts, err := json.Marshal(nodeKey.PubKey())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pubkey: %w", err)
	}

	tx.Sender = string(keybts)
	fmt.Println("tx hash", tx.Hash)
	fmt.Println("tx sender", tx.Sender)
	fmt.Println("tx signature", tx.Signature)
	fmt.Println("tx payload", tx.Payload)
	fmt.Println("tx payload type", tx.PayloadType)

	return tx, nil
}

func (c *Client) getAddress() (string, error) {
	if c.PrivateKey == nil {
		return "", fmt.Errorf("private key is nil")
	}

	return crypto.AddressFromPrivateKey(c.PrivateKey), nil
}
