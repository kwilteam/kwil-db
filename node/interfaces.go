package node

import (
	"context"

	ktypes "github.com/kwilteam/kwil-db/core/types"
	blockprocessor "github.com/kwilteam/kwil-db/node/block_processor"
	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

type ConsensusEngine interface {
	Role() types.Role // maybe: Role() (rol types.Role, power int64)
	InCatchup() bool

	AcceptProposal(height int64, blkID, prevBlkID types.Hash, leaderSig []byte, timestamp int64) bool
	NotifyBlockProposal(blk *ktypes.Block)

	AcceptCommit(height int64, blkID types.Hash, appHash types.Hash, leaderSig []byte) bool
	NotifyBlockCommit(blk *ktypes.Block, appHash types.Hash)

	NotifyACK(validatorPK []byte, ack types.AckRes)

	NotifyResetState(height int64, txIDs []types.Hash)

	NotifyDiscoveryMessage(validatorPK []byte, height int64)

	Start(ctx context.Context, proposerBroadcaster consensus.ProposalBroadcaster,
		blkAnnouncer consensus.BlkAnnouncer, ackBroadcaster consensus.AckBroadcaster,
		blkRequester consensus.BlkRequester, stateResetter consensus.ResetStateBroadcaster, discoveryBroadcaster consensus.DiscoveryReqBroadcaster, txBroadcaster blockprocessor.BroadcastTxFn) error

	CheckTx(ctx context.Context, tx *ktypes.Transaction) error

	ConsensusParams() *ktypes.ConsensusParams
	CancelBlockExecution(height int64, txIDs []types.Hash) error
}

type BlockProcessor interface {
	GetValidators() []*ktypes.Validator
	SubscribeValidators() <-chan []*ktypes.Validator
}

type SnapshotStore interface {
	Enabled() bool
	GetSnapshot(height uint64, format uint32) *snapshotter.Snapshot
	ListSnapshots() []*snapshotter.Snapshot
	LoadSnapshotChunk(height uint64, format uint32, chunk uint32) ([]byte, error)
}

type DB interface {
	sql.ReadTxMaker
}
