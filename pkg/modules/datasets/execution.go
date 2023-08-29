package datasets

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/crypto/addresses"
	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

/*
	This files contains logic for executing state changes against the database.
*/

// ExecutionResponse is the response from any interaction that modifies state.
type ExecutionResponse struct {
	// Fee is the amount of tokens spent on the execution
	Fee     *big.Int
	GasUsed int64
}

// Deploy deploys a database.
func (u *DatasetModule) Deploy(ctx context.Context, schema *engineTypes.Schema, tx *transactions.Transaction) (*ExecutionResponse, error) {
	price, err := u.PriceDeploy(ctx, schema)
	if err != nil {
		if price == nil {
			price = big.NewInt(0)
		}
		return resp(price), err
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return resp(price), err
	}

	senderPubKey, err := tx.GetSenderPubKey()
	// NOTE: This should never happen, since the transaction is already validated
	if err != nil {
		return resp(price), fmt.Errorf("failed to parse sender: %w", err)
	}

	identifier, err := getUserIdentifier(senderPubKey)
	if err != nil {
		return resp(price), fmt.Errorf("failed to get user identifier: %w", err)
	}

	_, err = u.engine.CreateDataset(ctx, schema, identifier)
	if err != nil {
		return resp(price), fmt.Errorf("failed to create dataset: %w", err)
	}

	return resp(price), nil
}

// Drop drops a database.
func (u *DatasetModule) Drop(ctx context.Context, dbid string, tx *transactions.Transaction) (*ExecutionResponse, error) {
	price, err := u.PriceDrop(ctx, dbid)
	if err != nil {
		if price == nil {
			price = big.NewInt(0)
		}
		return resp(price), err
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return resp(price), err
	}

	senderPubKey, err := tx.GetSenderPubKey()
	// NOTE: This should never happen, since the transaction is already validated
	if err != nil {
		return resp(price), fmt.Errorf("failed to parse sender: %w", err)
	}

	identifier, err := getUserIdentifier(senderPubKey)
	if err != nil {
		return resp(price), fmt.Errorf("failed to get user identifier: %w", err)
	}

	err = u.engine.DropDataset(ctx, dbid, identifier)
	if err != nil {
		return resp(price), fmt.Errorf("failed to drop dataset: %w", err)
	}

	return resp(price), nil
}

// Execute executes an action against a database.
func (u *DatasetModule) Execute(ctx context.Context, dbid string, action string, args [][]any, tx *transactions.Transaction) (*ExecutionResponse, error) {
	price, err := u.PriceExecute(ctx, dbid, action, args)
	if err != nil {
		if price == nil {
			price = big.NewInt(0)
		}
		return resp(price), err
	}

	err = u.compareAndSpend(ctx, price, tx)
	if err != nil {
		return resp(price), err
	}

	senderPubKey, err := tx.GetSenderPubKey()
	// NOTE: This should never happen, since the transaction is already validated
	if err != nil {
		return resp(price), fmt.Errorf("failed to parse sender: %w", err)
	}

	identifier, err := getUserIdentifier(senderPubKey)
	if err != nil {
		return resp(price), fmt.Errorf("failed to get user identifier: %w", err)
	}

	_, err = u.engine.Execute(ctx, dbid, action, args,
		engine.WithCaller(identifier),
	)
	if err != nil {
		return resp(price), fmt.Errorf("failed to execute action: %w", err)
	}

	return resp(price), nil
}

// compareAndSpend compares the calculated price to the transaction's fee, and spends the price if the fee is sufficient.
func (u *DatasetModule) compareAndSpend(ctx context.Context, price *big.Int, tx *transactions.Transaction) error {
	if tx.Body.Fee.Cmp(price) < 0 {
		return fmt.Errorf(`%w: fee %s is less than price %s`, ErrInsufficientFee, tx.Body.Fee.String(), price.String())
	}

	senderPubKey, err := tx.GetSenderPubKey()
	// NOTE: This should never happen, since the transaction is already validated
	if err != nil {
		return fmt.Errorf("failed to parse sender: %w", err)
	}

	return u.accountStore.Spend(ctx, &balances.Spend{
		AccountAddress: senderPubKey.Address().String(),
		Amount:         price,
		Nonce:          int64(tx.Body.Nonce),
	})
}

func resp(fee *big.Int) *ExecutionResponse {
	return &ExecutionResponse{
		Fee:     fee,
		GasUsed: 0,
	}
}

// getUserIdentifier gets an identifier for the user based on their public key
// it currently treats Ethereum as the default format for Secp256k1 keys, and
// NEAR as the default format for Ed25519 keys
// in the future, the defaults should probably be configurable, and functionality
// should be added to support other formats
func getUserIdentifier(pub crypto.PublicKey) (*addresses.KeyIdentifier, error) {
	var addressFormat addresses.AddressFormat
	switch pub.Type() {
	default: // this should never happen
		return nil, fmt.Errorf("unknown public key type: %s", pub.Type())
	case crypto.Secp256k1:
		addressFormat = addresses.AddressFormatEthereum
	case crypto.Ed25519:
		addressFormat = addresses.AddressFormatNEAR
	}

	return addresses.CreateKeyIdentifier(pub, addressFormat)
}
