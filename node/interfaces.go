package node

import (
	"context"

	"github.com/kwilteam/kwil-db/node/consensus"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

type ConsensusEngine interface {
	Role() types.Role

	AcceptProposal(height int64, blkID, prevBlkID types.Hash, leaderSig []byte, timestamp int64) bool
	NotifyBlockProposal(blk *types.Block)

	AcceptCommit(height int64, blkID types.Hash, appHash types.Hash, leaderSig []byte) bool
	NotifyBlockCommit(blk *types.Block, appHash types.Hash)

	NotifyACK(validatorPK []byte, ack types.AckRes)

	NotifyResetState(height int64)

	NotifyDiscoveryMessage(validatorPK []byte, height int64)

	Start(ctx context.Context, proposerBroadcaster consensus.ProposalBroadcaster,
		blkAnnouncer consensus.BlkAnnouncer, ackBroadcaster consensus.AckBroadcaster,
		blkRequester consensus.BlkRequester, stateResetter consensus.ResetStateBroadcaster, discoveryBroadcaster consensus.DiscoveryReqBroadcaster) error
}

type SnapshotStore interface {
	Enabled() bool
	GetSnapshot(height uint64, format uint32) *snapshotter.Snapshot
	ListSnapshots() []*snapshotter.Snapshot
	LoadSnapshotChunk(height uint64, format uint32, chunk uint32) ([]byte, error)
}

type DB interface {
	sql.TxMaker // for out-of-consensus writes e.g. setup and meta table writes
	sql.PreparedTxMaker
	sql.ReadTxMaker
	sql.SnapshotTxMaker
}
