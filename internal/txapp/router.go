// package tx_router routes transactions to the appropriate module(s)
package txapp

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/validators"
	"go.uber.org/zap"
)

// NewRouter creates a new router.
func NewRouter(db DatabaseEngine, acc AccountsStore, validators ValidatorStore, atomicCommitter AtomicCommitter, log log.Logger) *TxApp {
	return &TxApp{
		// TODO: set the eventstore and votestore dependencies
		Database:        db,
		Accounts:        acc,
		Validators:      validators,
		atomicCommitter: atomicCommitter,
		log:             log,
		mempool: &mempool{
			accountStore: acc,
			accounts:     make(map[string]*accounts.Account),
		},
	}
}

// TxApp maintains the state for Kwil's ABCI application.
// It is responsible for interpreting payload bodies and routing them properly.
// It also contains a mempool for uncommitted accounts, as well as pricing
// for transactions
type TxApp struct {
	Database       DatabaseEngine // tracks deployed schemas
	Accounts       AccountsStore  // accounts
	Validators     ValidatorStore // validators
	VoteStore      VoteStore      // tracks resolutions, their votes, manages expiration
	LocalValidator LocalValidator // information about the local validator
	NetworkInfo    NetworkInfo    // information about the network
	EventStore     EventStore     // tracks events, not part of consensus

	log log.Logger

	atomicCommitter AtomicCommitter
	mempool         *mempool
}

// Execute executes a transaction.  It will route the transaction to the
// appropriate module(s) and return the response.
func (r *TxApp) Execute(ctx TxContext, tx *transactions.Transaction) *TxResponse {
	route, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return txRes(nil, transactions.CodeInvalidTxType, fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String()))
	}

	r.log.Debug("executing transaction", zap.Any("tx", tx))

	return route.Execute(ctx, r, tx)
}

// Begin signals that a new block has begun.
func (r *TxApp) Begin(ctx context.Context, blockHeight int64) error {
	idempotencyKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(idempotencyKey, uint64(blockHeight))

	r.log.Debug("beginning block", zap.Int64("blockHeight", blockHeight))

	return r.atomicCommitter.Begin(ctx, idempotencyKey)
}

// Commit signals that a block has been committed.
// TODO: once we use postgres, this will no longer be applicable
// we will need a separate function for getting end results and committing
// Right now, Commit is called in FinalizeBlock in abci. However, it should
// be called in Commit.  The reason we can get away with this is because
// we rely on idempotency keys to ensure we don't double execute to a datastore.
// With Postgres, we will simply rely on its cross-schema
// transaction support.  Therefore, we should have another method here called
// GetEndResults.
// Commit also clears the mempool.
func (r *TxApp) Commit(ctx context.Context, blockHeight int64) (apphash []byte, validatorUpgrades []*types.Validator, err error) {
	// this would go in Commit
	defer r.mempool.reset()

	r.log.Debug("committing block", zap.Int64("blockHeight", blockHeight))

	// this would go in finalize block
	// run all approved votes, and delete from local store
	finalizedEvents, err := r.VoteStore.ProcessConfirmedResolutions(ctx)
	if err != nil {
		return nil, nil, err
	}

	// this would also go in finalize block
	for _, eventID := range finalizedEvents {
		err = r.EventStore.DeleteEvent(ctx, eventID)
		if err != nil {
			return nil, nil, err
		}
	}

	// expire votes
	// this would go in FinalizeBlock
	err = r.VoteStore.Expire(ctx, blockHeight)
	if err != nil {
		return nil, nil, err
	}

	// this would go in GetEndResults
	validators, err := r.Validators.Finalize(ctx)
	if err != nil {
		return nil, nil, err
	}

	// this would go in Commit
	idempotencyKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(idempotencyKey, uint64(blockHeight))

	// appHash would go in GetEndResults,
	// the commit would go in Commit
	appHash, err := r.atomicCommitter.Commit(ctx, idempotencyKey)
	if err != nil {
		return nil, nil, err
	}

	// this only updates an in-memory value. but it seems weird to me that the validator store needs to be aware
	// of the current block height and "keep it"
	// this would go in Commit
	r.Validators.UpdateBlockHeight(blockHeight)

	return appHash, validators, nil
}

// ApplyMempool applies the transactions in the mempool.
// If it returns an error, then the transaction is invalid.
func (r *TxApp) ApplyMempool(ctx context.Context, tx *transactions.Transaction) error {
	// check that payload type is valid
	_, ok := routes[tx.Body.PayloadType.String()]
	if !ok {
		return fmt.Errorf("unknown payload type: %s", tx.Body.PayloadType.String())
	}

	return r.mempool.applyTransaction(ctx, tx)
}

// GetAccount gets account info from either the mempool or the account store.
// It takes a flag to indicate whether it should check the mempool first.
func (r *TxApp) AccountInfo(ctx context.Context, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error) {
	var a *accounts.Account
	if getUncommitted {
		a, err = r.mempool.accountInfoSafe(ctx, acctID)
	} else {
		a, err = r.Accounts.GetAccount(ctx, acctID)
	}
	if err != nil {
		return nil, 0, err
	}

	return a.Balance, a.Nonce, nil
}

// ProposerTxs returns the transactions that the proposer should include in the
// next block.
func (r *TxApp) ProposerTxs(ctx context.Context) ([]*transactions.Transaction, error) {
	events, err := r.EventStore.GetEvents(ctx)
	if err != nil {
		return nil, err
	}

	account, err := r.Accounts.GetAccount(ctx, r.LocalValidator.Signer().Identity())
	if err != nil {
		return nil, err
	}

	tx, err := transactions.CreateTransaction(&transactions.ValidatorVoteBodies{
		Events: events,
	}, r.NetworkInfo.ChainID(), uint64(account.Nonce+1))
	if err != nil {
		return nil, err
	}

	err = tx.Sign(r.LocalValidator.Signer())
	if err != nil {
		return nil, err
	}

	return []*transactions.Transaction{tx}, nil
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
func (r *TxApp) Price(ctx TxContext, tx *transactions.Transaction) (*big.Int, error) {
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
	AccountReader
	Credit(ctx context.Context, acctID []byte, amt *big.Int) error
	Transfer(ctx context.Context, to, from []byte, amt *big.Int) error
	Spend(ctx context.Context, spend *accounts.Spend) error
}

// AccountReader is a datastore that can read accounts.
// It should not be used during block execution, since it does not read
// uncommitted accounts.
type AccountReader interface {
	// GetAccount gets an account from the datastore.
	// It should not be used during block execution, since it does not read
	// uncommitted accounts.
	GetAccount(ctx context.Context, acctID []byte) (*accounts.Account, error)
}

// ValidatorStore is a datastore that tracks validator information.
type ValidatorStore interface {
	Join(ctx context.Context, joiner []byte, power int64) error
	Leave(ctx context.Context, joiner []byte) error
	Approve(ctx context.Context, joiner, approver []byte) error
	Remove(ctx context.Context, target, validator []byte) error
	// Finalize is used at the end of block processing to retrieve the validator
	// updates to be provided to the consensus client for the next block. This
	// is not idempotent. The modules working list of updates is reset until
	// subsequent join/approves are processed for the next block.
	Finalize(ctx context.Context) ([]*validators.Validator, error) // end of block processing requires providing list of updates to the node's consensus client

	// Updates block height stored by the validator manager. Called in the abci Commit
	UpdateBlockHeight(blockHeight int64)

	// IsCurrent returns true if the validator is currently a validator.
	// It does not take into account uncommitted changes, but is thread-safe.
	IsCurrent(ctx context.Context, validator []byte) (bool, error)
}

// LocalValidator returns information about the local validator.
type LocalValidator interface {
	Signer() auth.Signer
}

// NetworkInfo contains information about the network.
type NetworkInfo interface {
	ChainID() string
}

// VoteStore is a datastore that tracks votes.
type VoteStore interface {
	// Approve approves a resolution.
	// If the resolution already includes a body, then it will return true.
	Approve(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error
	ContainsBody(ctx context.Context, resolutionID types.UUID) (bool, error)
	CreateResolution(ctx context.Context, event *types.VotableEvent, expiration int64) error
	Expire(ctx context.Context, blockheight int64) error
	UpdateVoter(ctx context.Context, identifier []byte, power int64) error
	// ProcessConfirmedResolutions processes all resolutions that have been confirmed.
	// It returns an array of the ID of the resolutions that were processed.
	ProcessConfirmedResolutions(ctx context.Context) ([]types.UUID, error)
	// AlreadyProcessed returns true if the resolution ID has been voted on (either succeeded, or expired).
	AlreadyProcessed(ctx context.Context, resolutionID types.UUID) (bool, error)
	// HasVoted returns true if the voter has voted on the resolution.
	HasVoted(ctx context.Context, resolutionID types.UUID, voter []byte) (bool, error)
}

// EventStore is a datastore that tracks events.
type EventStore interface {
	DeleteEvent(ctx context.Context, id types.UUID) error
	GetEvents(ctx context.Context) ([]*types.VotableEvent, error)
}

// AtomicCommitter is an interface for a struct that implements atomic commits across multiple stores
type AtomicCommitter interface {
	Begin(ctx context.Context, idempotencyKey []byte) error
	Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error)
}

// Broadcaster can broadcast transactions to the network.
type Broadcaster interface {
	BroadcastTx(ctx context.Context, tx []byte, sync uint8) (code uint32, txHash []byte, err error)
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
func (r *TxApp) checkAndSpend(ctx TxContext, tx *transactions.Transaction) (*big.Int, transactions.TxCode, error) {
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

	// check if the transaction consented to spending enough tokens
	if tx.Body.Fee.Cmp(amt) < 0 {
		return nil, transactions.CodeInsufficientFee, fmt.Errorf("transaction does not consent to spending enough tokens. transaction fee: %s, required fee: %s", tx.Body.Fee.String(), amt.String())
	}

	// spend the tokens
	err = r.Accounts.Spend(ctx.Ctx(), &accounts.Spend{
		AccountID: tx.Sender,
		Amount:    amt,
		Nonce:     int64(tx.Body.Nonce),
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
