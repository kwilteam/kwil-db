package txapp

import (
	"context"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/events"
	"github.com/kwilteam/kwil-db/internal/voting"
)

// Rebroadcaster is a service that marks events for rebroadcasting.
type Rebroadcaster interface {
	// MarkRebroadcast marks events for rebroadcasting.
	MarkRebroadcast(ctx context.Context, ids []types.UUID) error
}

// DB is the interface for the main SQL database.
type DB interface {
	sql.OuterTxMaker
	sql.ReadTxMaker
}

// package level funcs
// these can be overridden for testing
var (
	// getEvents gets all events, even if they have been
	// marked received
	getEvents = events.GetEvents
	// markReceived marks that an event from the local
	// validator has been received by the network.
	// This tells the event store to not re-broadcast the event,
	// but also to not delete it, as it may need to get the event
	// body in case it is a future block proposer.
	markReceived = events.MarkReceived
	// deleteEvent deletes an event. It will no longer
	// be broadcasted.
	deleteEvent = events.DeleteEvent

	// voting
	setVoterPower                    = voting.SetValidatorPower
	getAllVoters                     = voting.GetValidators
	getResolutionsByThresholdAndType = voting.GetResolutionsByThresholdAndType
	deleteResolutions                = voting.DeleteResolutions
	markProcessed                    = voting.MarkProcessed
	getExpired                       = voting.GetExpired
	requiredPower                    = voting.RequiredPower
	isProcessed                      = voting.IsProcessed
	resolutionContainsBody           = voting.ResolutionContainsBody
	getResolutionsByTypeAndProposer  = voting.GetResolutionIDsByTypeAndProposer
	createResolution                 = voting.CreateResolution
	approveResolution                = voting.ApproveResolution
	getVoterPower                    = voting.GetValidatorPower
	hasVoted                         = voting.HasVoted

	// account functions
	getAccount = accounts.GetAccount
	credit     = accounts.Credit
	spend      = accounts.Spend
	transfer   = accounts.Transfer
)
