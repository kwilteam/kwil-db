package abci

import (
	"encoding/json"

	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/abci/snapshots"

	abciTypes "github.com/cometbft/cometbft/abci/types"
)

func convertABCISnapshots(req *abciTypes.Snapshot) *snapshots.Snapshot {
	var metadata snapshots.SnapshotMetadata
	err := json.Unmarshal(req.Metadata, &metadata)
	if err != nil {
		return nil
	}

	snapshot := &snapshots.Snapshot{
		Height:     req.Height,
		Format:     req.Format,
		ChunkCount: req.Chunks,
		Hash:       req.Hash,
		Metadata:   metadata,
	}
	return snapshot
}

func convertToABCISnapshot(snapshot *snapshots.Snapshot) (*abciTypes.Snapshot, error) {
	metadata, err := json.Marshal(snapshot.Metadata)
	if err != nil {
		return nil, err
	}

	return &abciTypes.Snapshot{
		Height:   snapshot.Height,
		Format:   snapshot.Format,
		Chunks:   snapshot.ChunkCount,
		Hash:     snapshot.Hash,
		Metadata: metadata,
	}, nil
}

func abciStatus(status snapshots.Status) abciTypes.ResponseApplySnapshotChunk_Result {
	switch status {
	case snapshots.ACCEPT:
		return abciTypes.ResponseApplySnapshotChunk_ACCEPT
	case snapshots.REJECT:
		return abciTypes.ResponseApplySnapshotChunk_REJECT_SNAPSHOT
	case snapshots.RETRY:
		return abciTypes.ResponseApplySnapshotChunk_RETRY
	default:
		return abciTypes.ResponseApplySnapshotChunk_UNKNOWN
	}
}

// groupTransactions groups the transactions by sender.
func groupTxsBySender(txns [][]byte) (map[string][]*transactions.Transaction, error) {
	grouped := make(map[string][]*transactions.Transaction)
	for _, tx := range txns {
		t := &transactions.Transaction{}
		err := t.UnmarshalBinary(tx)
		if err != nil {
			return nil, err
		}
		key := string(t.Sender)
		grouped[key] = append(grouped[key], t)
	}
	return grouped, nil
}

// nonceList is for debugging
func nonceList(txns []*transactions.Transaction) []uint64 {
	nonces := make([]uint64, len(txns))
	for i, tx := range txns {
		nonces[i] = tx.Body.Nonce
	}
	return nonces
}
