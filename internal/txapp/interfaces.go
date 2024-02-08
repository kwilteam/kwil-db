package txapp

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/accounts"
	engineTypes "github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/validators"
)

// file defines interface dependencies for the txapp package.

// ExecutionEngine is a database that can handle deployments, executions, etc.
type ExecutionEngine interface {
	CreateDataset(ctx context.Context, db sql.DB, schema *engineTypes.Schema, caller []byte) (err error)
	DeleteDataset(ctx context.Context, db sql.DB, dbid string, caller []byte) error
	Execute(ctx context.Context, db sql.DB, data *engineTypes.ExecutionData) (*sql.ResultSet, error)
}

// AccountsStore is a datastore that can handle accounts.
type AccountsStore interface {
	AccountReader
	Credit(ctx context.Context, tx sql.DB, acctID []byte, amt *big.Int) error
	Transfer(ctx context.Context, tx sql.DB, to, from []byte, amt *big.Int) error
	Spend(ctx context.Context, tx sql.DB, spend *accounts.Spend) error
}

// AccountReader is a datastore that can read accounts.
// It should not be used during block execution, since it does not read
// uncommitted accounts.
type AccountReader interface {
	// GetAccount gets an account from the datastore.
	// It should not be used during block execution, since it does not read
	// uncommitted accounts.
	GetAccount(ctx context.Context, tx sql.DB, acctID []byte) (*accounts.Account, error)
}

// ValidatorStore is a datastore that tracks validator information.
type ValidatorStore interface {
	IsValidatorChecker
	Join(ctx context.Context, tx sql.DB, joiner []byte, power int64) error
	Leave(ctx context.Context, tx sql.DB, joiner []byte) error
	Approve(ctx context.Context, tx sql.DB, joiner, approver []byte) error
	Remove(ctx context.Context, tx sql.DB, target, validator []byte) error
	// Finalize is used at the end of block processing to retrieve the validator
	// updates to be provided to the consensus client for the next block. This
	// is not idempotent. The modules working list of updates is reset until
	// subsequent join/approves are processed for the next block.
	Finalize(ctx context.Context, tx sql.DB) ([]*validators.Validator, error) // end of block processing requires providing list of updates to the node's consensus client

	// Updates block height stored by the validator manager. Called in the abci Commit
	UpdateBlockHeight(blockHeight int64)

	// StateHash returns a hash representing the current state of the validator store
	StateHash() []byte
	// GenesisInit configures the initial set of validators for the genesis
	// block. This is only called once for a new chain.
	GenesisInit(ctx context.Context, tx sql.DB, vals []*validators.Validator, blockHeight int64) error

	// Update updates the power of a validator.
	Update(ctx context.Context, tx sql.DB, validator []byte, newPower int64) error

	// CurrentSet returns the current validator set.
	CurrentSet(ctx context.Context, tx sql.DB) ([]*validators.Validator, error)
}

// part of the validator store, but split up to delineate between TxApp and
// mempool dependencies
type IsValidatorChecker interface {
	// IsCurrent returns true if the validator is currently a validator.
	// It does not take into account uncommitted changes, but is thread-safe.
	IsCurrent(ctx context.Context, tx sql.DB, validator []byte) (bool, error)
}

// LocalValidator returns information about the local validator.
type LocalValidator interface {
	Signer() *auth.Ed25519Signer
}

// NetworkInfo contains information about the network.
type NetworkInfo interface {
	ChainID() string
}

// VoteStore is a datastore that tracks votes.
type VoteStore interface {
	// Approve approves a resolution.
	// If the resolution already includes a body, then it will return true.
	Approve(ctx context.Context, tx sql.DB, resolutionID types.UUID, expiration int64, from []byte) error
	// ContainsBodyOrFinished returns true if (any of the following are true):
	// 1. the resolution has a body
	// 2. the resolution has expired
	// 3. the resolution has been approved
	ContainsBodyOrFinished(ctx context.Context, tx sql.DB, resolutionID types.UUID) (bool, error)
	CreateResolution(ctx context.Context, tx sql.DB, event *types.VotableEvent, expiration int64) error
	Expire(ctx context.Context, tx sql.DB, blockheight int64) error
	UpdateVoter(ctx context.Context, tx sql.DB, identifier []byte, power int64) error
	// ProcessConfirmedResolutions processes all resolutions that have been confirmed.
	// It returns an array of the ID of the resolutions that were processed.
	ProcessConfirmedResolutions(ctx context.Context, tx sql.DB) ([]types.UUID, error)
	// HasVoted returns true if the voter has voted on the resolution.
	HasVoted(ctx context.Context, tx sql.DB, resolutionID types.UUID, voter []byte) (bool, error)
}

// EventStore is a datastore that tracks events.
type EventStore interface {
	// DeleteEvent deletes an event. It will not longer
	// be broadcasted
	DeleteEvent(ctx context.Context, db sql.DB, id types.UUID) error
	// GetEvents gets all events, even if they have been
	// marked received
	GetEvents(ctx context.Context, db sql.DB) ([]*types.VotableEvent, error)
	// MarkReceived marks that an event from the local
	// validator has been received by the network.
	// This tells the event store to not re-broadcast the event,
	// but also to not delete it, as it may need to get the event
	// body in case it is a future block proposer.
	MarkReceived(ctx context.Context, db sql.DB, id types.UUID) error
}

// DB is the interface for the main SQL database.
type DB interface {
	sql.OuterTxMaker
	sql.ReadTxMaker
}
