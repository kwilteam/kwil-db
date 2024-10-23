package dummyce

import (
	"context"
	"fmt"
	"log"
	"p2p/node/types"
	"sync"
	"sync/atomic"
	"time"
)

type blkCommit struct {
	height int64
	hash   types.Hash
}

type blkProp struct {
	height int64
	hash   types.Hash
	blk    *types.Block
	resCb  func(ack bool, appHash types.Hash) error
}

type ackFrom struct {
	fromPubKey []byte
	res        types.AckRes
}

type blkResult struct {
	commit func() error

	appHash types.Hash
	txRes   []types.TxResult
	// other updates for next block...
}

type Engine struct {
	bki types.BlockStore
	txi types.TxIndex
	mp  types.MemPool

	exec types.Execution

	wg     sync.WaitGroup
	leader atomic.Bool
	mined  chan *types.QualifiedBlock

	mtx        sync.RWMutex
	lastCommit blkCommit
	validators [][]byte
	proposed   *blkProp
	prepared   *blkResult
	// as leader, we collect acks for the proposed block. after ack threshold,
	// we commit and clear proposed/prepared for next block
	acks []ackFrom
}

func New(bs types.BlockStore, txi types.TxIndex, mp types.MemPool) *Engine {
	return &Engine{
		bki:   bs,
		txi:   txi,
		mp:    mp,
		mined: make(chan *types.QualifiedBlock, 1),
	}
}

func (ce *Engine) Start(ctx context.Context) error {
	const blkInterval = 4 * time.Second
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		ce.mine(ctx, blkInterval)
	}()

	ce.wg.Wait()

	<-ctx.Done()

	return nil
}

// CommitBlock reports a full block to commit. This would be used when:
//  1. retrieval of a block following in an announcement for a new+next block
//  2. iterative block retrieval in catch-up / sync
func (ce *Engine) CommitBlock(blk *types.Block, appHash types.Hash) error {
	height := blk.Header.Height
	if ce.proposed == nil {
		// execute and commit if it is next
		if height != ce.lastCommit.height+1 {
			return fmt.Errorf("block at height %d does not follow %d", height, ce.lastCommit.height)
		}
		// TODO: execute and commit
		return nil
	}

	if ce.proposed.height != height {
		return fmt.Errorf("block at height %d does not match existing proposed block at %d", height, ce.proposed.height)
	}

	blkHash := blk.Header.Hash()
	if ce.proposed.hash != blkHash {
		return fmt.Errorf("block at height %d with hash %v does not match hash of existing proposed block %d",
			blkHash, ce.proposed.hash, height)
	}

	// TODO: flag OK to commit ce.proposed.blk (with expected apphash)

	return nil
}

func (ce *Engine) confirmBlkTxns(blk *types.Block) {
	blkHash := blk.Header.Hash()
	height := blk.Header.Height

	log.Printf("confirming %d transactions in block %d (%v)", len(blk.Txns), height, blkHash)
	for _, txn := range blk.Txns {
		txHash := types.HashBytes(txn)
		ce.txi.Store(txHash, txn) // add to tx index
		ce.mp.Store(txHash, nil)  // rm from mempool
	}

	rawBlk := types.EncodeBlock(blk)
	ce.bki.Store(blkHash, height, rawBlk)
}
