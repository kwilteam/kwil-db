package txapp

import (
	"context"
	"kwil/node/types/sql"
	"kwil/node/voting"
	"kwil/types"
	"math/big"
)

type Accounts interface {
	Spend(ctx context.Context, tx sql.Executor, acctID []byte, amount *big.Int, nonce int64) error
	Credit(ctx context.Context, tx sql.Executor, acctID []byte, amount *big.Int) error
	Transfer(ctx context.Context, tx sql.Executor, from, to []byte, amount *big.Int) error
	GetAccount(ctx context.Context, tx sql.Executor, acctID []byte) (*types.Account, error)
	ApplySpend(ctx context.Context, tx sql.Executor, acctID []byte, amount *big.Int, nonce int64) error
	Commit() error
}

type Validators interface {
	SetValidatorPower(ctx context.Context, tx sql.Executor, pubKey []byte, power int64) error
	GetValidatorPower(ctx context.Context, tx sql.Executor, pubKey []byte) (int64, error)
	GetValidators() ([]*types.Validator, error)
	Commit() error
}

// Rebroadcaster is a service that marks events for rebroadcasting.
type Rebroadcaster interface {
	// MarkRebroadcast marks events for rebroadcasting.
	MarkRebroadcast(ctx context.Context, ids []*types.UUID) error
}

// DB is the interface for the main SQL database. All queries must be executed
// from within a transaction. A DB can create read transactions or the special
// two-phase outer write transaction.
type DB interface {
	sql.PreparedTxMaker
	sql.ReadTxMaker
	sql.SnapshotTxMaker
}

var (
	// getEvents gets all events, even if they have been
	// marked received
	getEvents = voting.GetEvents

	// deleteEvent deletes an event. It will no longer
	// be broadcasted.
	deleteEvent  = voting.DeleteEvent
	deleteEvents = voting.DeleteEvents

	// voting
	getResolutionsByThresholdAndType = voting.GetResolutionsByThresholdAndType // called from RW consensus tx
	deleteResolutions                = voting.DeleteResolutions
	markProcessed                    = voting.MarkProcessed
	getExpired                       = voting.GetExpired
	requiredPower                    = voting.RequiredPower
	getResolutionsByTypeAndProposer  = voting.GetResolutionIDsByTypeAndProposer
	createResolution                 = voting.CreateResolution
	approveResolution                = voting.ApproveResolution
	resolutionExists                 = voting.ResolutionExists
	resolutionByID                   = voting.GetResolutionInfo
	deleteResolution                 = voting.DeleteResolution
)
