// package tx_router routes transactions to the appropriate module(s)
package txrouter

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// Router routes incoming transactions to the appropriate module(s)
// It is capable of sending to the database, spending, adding/removing
// validators, etc.
type Router struct {
	Database   DatabaseEngine
	Accounts   AccountsStore
	Validators ValidatorStore
}

// Execute executes a transaction.  It will route the transaction to the
// appropriate module(s) and return the response.
func (r *Router) Execute(ctx context.Context, tx *transactions.Transaction) *TxResponse {
	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return txRes(nil, transactions.CodeInvalidTxType, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String()))
	}

	return route.Execute(ctx, r, tx)
}

// Begin signals that a new block has begun.
func (r *Router) Begin(ctx context.Context, blockHeight int64) error {
	return nil
}

// Commit signals that a block has been committed.
// TODO: once we use postgres, this will no longer be applicable
// we will need a separate function for getting end results and committing
// Right now, Commit is called in FinalizeBlock in abci. However, it should
// be called in Commit.  The reason we can get away with this is because
// we rely on idempotency keys to ensure we don't double execute to a datastore.
// With Postgres, we will simply rely on its cross-schema
// transaction support.
func (r *Router) Commit(ctx context.Context, blockHeight int64) (apphash []byte, validatorUpgrades []*types.Validator, err error) {
	return nil, nil, nil
}

// TxResponse is the response from a transaction.
// It contains information about the transaction, such as the amount spent.
type TxResponse struct {
	// ResponseCode is the response code from the transaction
	ResponseCode transactions.TxCode

	// Spend is the amount of tokens spent by the transaction
	Spend int64

	// Error is the error returned by the transaction, if any
	Error error
}

// Price estimates the price of a transaction.
// It returns the estimated price in tokens.
func (r *Router) Price(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return nil, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}

	return route.Price(ctx, r, tx)
}

// DatabaseEngine is a database that can handle deployments, executions, etc.
type DatabaseEngine interface {
	CreateDataset(ctx context.Context, schema *engineTypes.Schema, caller []byte) (err error)
	DeleteDataset(ctx context.Context, dbid string, caller []byte) error
	Execute(ctx context.Context, data *engineTypes.ExecutionData) (*sql.ResultSet, error)
}

// AccountsStore is a datastore that can handle accounts.
type AccountsStore interface {
	GetAccount(ctx context.Context, acctID []byte) (*accounts.Account, error)
	Credit(ctx context.Context, acctID []byte, amt *big.Int) error
	Transfer(ctx context.Context, to, from []byte, amt *big.Int) error
	Spend(ctx context.Context, spend *accounts.Spend) error
}

// ValidatorStore is a datastore that tracks validator information.
type ValidatorStore interface {
	Join(ctx context.Context, joiner []byte, power int64) error
	Leave(ctx context.Context, joiner []byte) error
	Approve(ctx context.Context, joiner, approver []byte) error
	Remove(ctx context.Context, target, validator []byte) error
}

// AtomicCommitter is an interface for a struct that implements atomic commits across multiple stores
type AtomicCommitter interface {
	Begin(ctx context.Context, idempotencyKey []byte) error
	Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error)
}

// checkAndSpend checks the price of a transaction.
// it returns the price it will cost to execute the transaction.
// if the transaction does not have enough tokens to pay for the transaction,
// it will return an error.
// if the caller does not have enough tokens to pay for the transaction,
// it will return an error.
// if the transaction does not have the correct nonce, it will return an error.
// it will spend the tokens if the caller has enough tokens.
// It also returns an error code.
// if we allow users to implement their own routes, this function will need to
// be exported.
func (r *Router) checkAndSpend(ctx context.Context, tx *transactions.Transaction) (*big.Int, transactions.TxCode, error) {
	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return nil, transactions.CodeInvalidTxType, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}

	amt, err := route.Price(ctx, r, tx)
	if err != nil {
		return nil, transactions.CodeUnknownError, err
	}

	if amt.Cmp(tx.Body.Fee) < 0 {
		return nil, transactions.CodeInsufficientFee, fmt.Errorf("transaction fee is too low: %s", amt.String())
	}

	acct, err := r.Accounts.GetAccount(ctx, tx.Sender)
	if err != nil {
		return nil, transactions.CodeUnknownError, err
	}

	// check if the account has enough tokens to pay for the transaction
	if acct.Balance.Cmp(amt) < 0 {
		return nil, transactions.CodeInsufficientBalance, fmt.Errorf("account %s does not have enough tokens to pay for transaction. account balance: %s, required balance: %s", tx.Sender, acct.Balance.String(), amt.String())
	}

	// check if the transaction consented to spending enough tokens
	if tx.Body.Fee.Cmp(amt) < 0 {
		return nil, transactions.CodeInsufficientFee, fmt.Errorf("transaction does not consent to spending enough tokens. transaction fee: %s, required fee: %s", tx.Body.Fee.String(), amt.String())
	}

	// check the nonce
	// this is somewhat redundant with the account store, but we can add the correct
	// error code here
	if acct.Nonce != int64(tx.Body.Nonce)+1 {
		return nil, transactions.CodeInvalidNonce, fmt.Errorf("invalid nonce. account nonce: %d, transaction nonce: %d", acct.Nonce, tx.Body.Nonce)
	}

	// spend the tokens
	err = r.Accounts.Spend(ctx, &accounts.Spend{
		AccountID: tx.Sender,
		Amount:    amt,
		Nonce:     int64(tx.Body.Nonce) + 1,
	})
	if err != nil {
		return nil, transactions.CodeUnknownError, err
	}

	return amt, transactions.CodeOk, nil
}

// txRes wraps a spend, tx code, and error into a tx response.
// the spend amount is included because an error can occur after the tokens
// are spent.
func txRes(spend *big.Int, code transactions.TxCode, err error) *TxResponse {
	if spend == nil {
		spend = big.NewInt(0)
	}

	return &TxResponse{
		ResponseCode: code,
		Spend:        spend.Int64(),
		Error:        err,
	}
}
