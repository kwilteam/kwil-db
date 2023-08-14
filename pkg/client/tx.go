package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

// Transaction signed by the client
func (c *Client) newTx(ctx context.Context, data transactions.Payload) (*transactions.Transaction, error) {
	if c.PrivateKey == nil {
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
	err = tx.Sign(c.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// Tx Signed by the Validator Node
// TODO: this needs to be updated once validator store (containing the validator payload) is merged in
func (c *Client) NewNodeTx(ctx context.Context, payloadType transactions.PayloadType, data any, privKey string) (*transactions.TransactionStatus, error) {
	panic("implement me")
	/*
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

		address := nodeKey.PubKey().Address().String()
		acc, err := c.client.GetAccount(ctx, address)
		if err != nil {
			acc = &balances.Account{
				Address: address,
				Nonce:   0,
				Balance: big.NewInt(0),
			}
		}

		// build transaction
		tx := kTx.NewTx(payloadType, bts, acc.Nonce+1)
		// sign transaction
		tx.Fee = "0"

		hash := tx.GenerateHash()
		sign, err := nodeKey.Sign(hash)
		if err != nil {
			return nil, fmt.Errorf("failed to sign tx: %w", err)
		}

		var keytype crypto.SignatureType
		if nodeKey.Type() == "ed25519" {
			keytype = crypto.SIGNATURE_TYPE_ED25519
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
	*/
}

func (c *Client) getAddress() (string, error) {
	if c.PrivateKey == nil {
		return "", fmt.Errorf("private key is nil")
	}

	return c.PrivateKey.PubKey().Address().String(), nil
}
