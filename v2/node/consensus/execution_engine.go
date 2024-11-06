package consensus

import (
	"context"
	"crypto/sha256"

	"kwil/node/types"
	ktypes "kwil/types"
)

// potentially our txApp
// - ConsensusTx
// 		- ExecuteTx(ctx, tx, nestedTx) -> TxResult
// - CommitTx
// - RollbackTx
// Cancellable context for tx execution

// Block Atomicity

// placeholder for our pg module constructs
// state.json will save the last commit info to tag the app state
// therefore blockExecutor should only track the intermediate state changes
// resulting from the txs within a block
type blockExecutor struct {
	changesets []types.Hash // changesets for each tx in a block, commit will give a hash of these changesets in a deterministic order
}

//	func (be *blockExecutor) BeginBlock() error {
//		return nil
//	}
func newBlockExecutor() *blockExecutor {
	return &blockExecutor{}
}

func (be *blockExecutor) Execute(_ context.Context, tx []byte) ktypes.TxResult {
	hash, _ := types.NewHashFromBytes(tx) // TODO: may also include the txresult hash
	be.changesets = append(be.changesets, hash)
	return ktypes.TxResult{
		Code: 0,
		Log:  "success" + hash.String(),
	}
}

// Precommit gives a deterministic hash based on the changesets resulting from the txs in a block
func (be *blockExecutor) Precommit() (types.Hash, error) {
	hasher := sha256.New()
	for _, changeset := range be.changesets {
		hasher.Write(changeset[:])
	}
	return types.Hash(hasher.Sum(nil)), nil
}

func (be *blockExecutor) Commit(commitFn func() error) error {
	// updates to the state should be done in the commitFn
	if err := commitFn(); err != nil {
		return err
	}

	be.changesets = nil
	return nil
}

// TODO: not much of any use yet.
func (be *blockExecutor) Rollback() error {
	be.changesets = nil
	return nil
}
