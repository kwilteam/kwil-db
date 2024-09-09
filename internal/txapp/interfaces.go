package txapp

import (
	"context"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/voting"
)

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

type Snapshotter interface {
	// CreateSnapshot creates a snapshot of the current state.
	CreateSnapshot(ctx context.Context, height uint64, snapshotID string) error

	// IsSnapshotDue returns true if a snapshot is due at the given height.
	IsSnapshotDue(height uint64) bool
}

// package level funcs
// these can be overridden for testing
var (
	// getEvents gets all events, even if they have been
	// marked received
	getEvents = voting.GetEvents

	// deleteEvent deletes an event. It will no longer
	// be broadcasted.
	deleteEvent  = voting.DeleteEvent
	deleteEvents = voting.DeleteEvents

	// voting
	setVoterPower                    = voting.SetValidatorPower
	getAllVoters                     = voting.GetValidators
	getResolutionsByThresholdAndType = voting.GetResolutionsByThresholdAndType // called from RW consensus tx
	deleteResolutions                = voting.DeleteResolutions
	markProcessed                    = voting.MarkProcessed
	getExpired                       = voting.GetExpired
	requiredPower                    = voting.RequiredPower
	getResolutionsByTypeAndProposer  = voting.GetResolutionIDsByTypeAndProposer
	createResolution                 = voting.CreateResolution
	approveResolution                = voting.ApproveResolution
	getVoterPower                    = voting.GetValidatorPower
	// resolutionExists                 = voting.ResolutionExists
	resolutionByID   = voting.GetResolutionInfo
	deleteResolution = voting.DeleteResolution

	// account functions
	getAccount = accounts.GetAccount
	credit     = accounts.Credit
	spend      = accounts.Spend
	applySpend = accounts.ApplySpend
	transfer   = accounts.Transfer
)
