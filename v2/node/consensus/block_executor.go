package consensus

import (
	"fmt"
	"p2p/node/types"
)

// Block processing methods

func (ce *ConsensusEngine) validateBlock(blk *types.Block) error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

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

func (ce *ConsensusEngine) executeBlock() error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	var txResults []txResult
	// Execute the block and return the appHash and store the txResults
	for _, tx := range ce.state.blkProp.blk.Txns {
		res, err := ce.blockExecutor.Execute(tx)
		if err != nil {
			return err
		}
		txResults = append(txResults, res)
	}

	// Calculate the appHash (dummy for now)
	appHash, err := types.NewHashFromString(string(ce.state.blkProp.blk.Header.PrevAppHash.String() + "random"))
	if err != nil {
		return err
	}

	ce.state.blockRes = &blockResult{
		txResults: txResults,
		appHash:   appHash,
	}

	return nil
}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (ce *ConsensusEngine) commit() error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// TODO: Lock mempool and update the mempool to remove the transactions in the block
	// Mempool should not receive any new transactions until this Commit is done as
	// we are updating the state and the tx checks should be done against the new state.

	// Add the block to the blockstore
	rawBlk := types.EncodeBlock(ce.state.blkProp.blk)
	ce.blockStore.Store(ce.state.blkProp.blkHash, ce.state.blkProp.height, rawBlk)

	// Commit the block to the postgres database
	if err := ce.blockExecutor.Commit(); err != nil {
		return err
	}

	// Update any other internal states like apphash and height to chain state and commit again

	// Add transactions to the txIndexer and remove them from the mempool
	for _, txn := range ce.state.blkProp.blk.Txns {
		txHash := types.HashBytes(txn)
		ce.indexer.Store(txHash, txn)
		ce.mempool.Store(txHash, nil)
	}

	return nil
}

func (ce *ConsensusEngine) nextState() {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	ce.state.lc = &lastCommit{
		height:  ce.state.blkProp.height,
		blkHash: ce.state.blkProp.blkHash,
		appHash: ce.state.blockRes.appHash,
	}

	ce.state.blkProp = nil
	ce.state.blockRes = nil

	ce.state.votes = make(map[string]*vote)
	ce.state.processedVotes = make(map[string]*vote)
	ce.state.acks = 0
	ce.state.nacks = 0
}

// TODO: not needed.
func tempBlock() *types.Block {
	return &types.Block{
		Header: &types.BlockHeader{
			Height: 1,
		},
		Txns: [][]byte{},
	}
}
