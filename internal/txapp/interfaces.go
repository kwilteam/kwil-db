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
	IsValidatorChecker
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
}

// part of the validator store, but split up to delineate between TxApp and
// mempool dependencies
type IsValidatorChecker interface {
	// IsCurrent returns true if the validator is currently a validator.
	// It does not take into account uncommitted changes, but is thread-safe.
	IsCurrent(ctx context.Context, validator []byte) (bool, error)
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
	Approve(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error
	// ContainsBodyOrFinished returns true if (any of the following are true):
	// 1. the resolution has a body
	// 2. the resolution has expired
	// 3. the resolution has been approved
	ContainsBodyOrFinished(ctx context.Context, resolutionID types.UUID) (bool, error)
	CreateResolution(ctx context.Context, event *types.VotableEvent, expiration int64) error
	Expire(ctx context.Context, blockheight int64) error
	UpdateVoter(ctx context.Context, identifier []byte, power int64) error
	// ProcessConfirmedResolutions processes all resolutions that have been confirmed.
	// It returns an array of the ID of the resolutions that were processed.
	ProcessConfirmedResolutions(ctx context.Context) ([]types.UUID, error)
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
