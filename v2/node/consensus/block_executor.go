package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"p2p/node/types"
	"path/filepath"
)

var (
	dirtyHash = types.HashBytes([]byte("0x42"))
)

// Block processing methods
func (ce *ConsensusEngine) validateBlock(blk *types.Block) error {
	// Validate if this is the correct block proposal to be processed.
	if blk.Header.Version != types.BlockVersion {
		return fmt.Errorf("block version mismatch, expected %d, got %d", types.BlockVersion, blk.Header.Version)
	}

	if ce.state.lc.height+1 != blk.Header.Height {
		return fmt.Errorf("block proposal for height %d does not follow %d", blk.Header.Height, ce.state.lc.height)
	}

	if ce.state.lc.blkHash != blk.Header.PrevHash {
		return fmt.Errorf("prevBlockHash mismatch, expected %v, got %v", ce.state.lc.blkHash, blk.Header.PrevHash)
	}

	if blk.Header.PrevAppHash != ce.state.lc.appHash {
		return fmt.Errorf("apphash mismatch, expected %v, got %v", ce.state.lc.appHash, blk.Header.PrevAppHash)
	}

	if blk.Header.NumTxns != uint32(len(blk.Txns)) {
		return fmt.Errorf("transaction count mismatch, expected %d, got %d", blk.Header.NumTxns, len(blk.Txns))
	}

	// Verify the merkle root of the block transactions
	merkleRoot := blk.MerkleRoot()
	if merkleRoot != blk.Header.MerkleRoot {
		return fmt.Errorf("merkleroot mismatch, expected %v, got %v", merkleRoot, blk.Header.MerkleRoot)
	}

	// Verify other stuff such as validatorsetHash, signature of the block etc.
	return nil
}

// TODO: need to stop execution when we receive a rollback signal of some sort.
func (ce *ConsensusEngine) executeBlock() error {
	var txResults []types.TxResult

	ctx, cancel := context.WithCancel(context.Background())
	ce.state.cancelFunc = cancel

	// TODO: test resetState
	// Execute the block and return the appHash and store the txResults
	for _, tx := range ce.state.blkProp.blk.Txns {
		res := ce.blockExecutor.Execute(ctx, tx) // TODO: this execute function should be context cancellable
		txResults = append(txResults, res)
	}

	ce.state.appState.Height = ce.state.blkProp.height
	ce.state.appState.AppHash = dirtyHash

	// Calculate the appHash (dummy for now)
	// Precommit equivalent
	cHash, err := ce.blockExecutor.Precommit()
	if err != nil {
		ce.log.Error("Failed to precommit the block tx", "err", err)
		return err
	}

	// Calculate the new apphash by hashing the previous apphash and the changeset hash
	appHash := sha256.Sum256(append(ce.state.blkProp.blk.Header.PrevAppHash[:], cHash[:]...))

	ce.state.blockRes = &blockResult{
		txResults: txResults,
		appHash:   appHash,
	}

	return nil
}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (ce *ConsensusEngine) commit() error {
	// TODO: Lock mempool and update the mempool to remove the transactions in the block
	// Mempool should not receive any new transactions until this Commit is done as
	// we are updating the state and the tx checks should be done against the new state.

	// Add the block to the blockstore
	ce.blockStore.Store(ce.state.blkProp.blk, ce.state.blockRes.appHash)
	ce.blockStore.StoreResults(ce.state.blkProp.blk.Header.Hash(), ce.state.blockRes.txResults)

	// Commit the block to the postgres database
	if err := ce.blockExecutor.Commit(ce.persistAppState); err != nil {
		return err
	}

	// Update any other internal states like apphash and height to chain state and commit again
	ce.state.appState.AppHash = ce.state.blockRes.appHash
	ce.persistAppState()

	// remove transactions from the mempool
	for _, txn := range ce.state.blkProp.blk.Txns {
		txHash := types.HashBytes(txn)
		// ce.indexer.Store(txHash, txn)
		ce.mempool.Store(txHash, nil)
	}

	ce.log.Info("Committed Block", "height", ce.state.blkProp.blk.Header.Height,
		"hash", ce.state.blkProp.blk.Header.Hash(), "appHash", ce.state.blockRes.appHash)
	return nil
}

func (ce *ConsensusEngine) resetBlockProp(rstMsg *resetState) {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// If we are currently executing any transactions corresponding to the blk at height +1
	// 1. Cancel the execution context -> so that the transactions stop
	// 2. Rollback the consensus tx
	// 3. Reset the blkProp and blockRes
	// 4. This should never happen after the commit phase, (blk should have never made it to the blockstore)

	// Ensure that the block is not committed yet. Else, we should just ignore the reset message (potentially a stale one)
	// TODO: this should probably be checked before accepting this message
	_, _, _, err := ce.blockStore.GetByHeight(rstMsg.height)
	if err != nil {
		if !errors.Is(err, types.ErrNotFound) {
			ce.log.Error("Error fetching block from blockstore", "height", rstMsg.height, "error", err)
		}
		// nothing to do here as the block is already committed
		return
	}

	if ce.state.lc.height == rstMsg.height {
		if ce.state.blkProp != nil {
			// first cancel the context
			ce.state.cancelFunc()
			// rollback the pg tx
			// ce.state.consensusTx.Rollback()

			// reset the blkProp and blockRes
			ce.resetState()
		}
	}
}

func (ce *ConsensusEngine) nextState() {
	ce.state.lc = &lastCommit{
		height:  ce.state.blkProp.height,
		blkHash: ce.state.blkProp.blkHash,
		appHash: ce.state.blockRes.appHash,
		blk:     ce.state.blkProp.blk,
	}

	ce.resetState()
}

func (ce *ConsensusEngine) resetState() {
	ce.state.blkProp = nil
	ce.state.blockRes = nil
	ce.state.votes = make(map[string]*vote)

	// reset the ctx and tx
	ce.state.cancelFunc = nil
	// ce.state.tx = nil
}

// temporary placeholder as this will be in the PG chainstate in future (as was in previous kwil implementations)
type appState struct {
	Height  int64      `json:"height"`
	AppHash types.Hash `json:"app_hash"`
}

func (ce *ConsensusEngine) persistAppState() error {
	bts, err := json.MarshalIndent(ce.state.appState, "", "  ")
	if err != nil {
		ce.log.Errorf("Error marshalling appstate: %v", err)
		return err // fatal or warn?
	}
	return os.WriteFile(ce.stateFile(), bts, 0644)
}

func (ce *ConsensusEngine) loadAppState() (*appState, error) {
	bts, err := os.ReadFile(ce.stateFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &appState{}, nil
		}
		return nil, fmt.Errorf("error reading appstate file: %w", err)
	}
	var state appState
	if err := json.Unmarshal(bts, &state); err != nil {
		return nil, fmt.Errorf("error unmarshalling appstate: %w", err)
	}
	return &state, nil
}

func (ce *ConsensusEngine) stateFile() string {
	return filepath.Join(ce.dir, "state.json")
}

func LoadState(filename string) (int64, types.Hash) {
	state := &appState{}
	bts, err := os.ReadFile(filename)
	if err != nil {
		return 0, types.Hash{}
	}
	if err := json.Unmarshal(bts, state); err != nil {
		return 0, types.Hash{}
	}
	return state.Height, state.AppHash
}
