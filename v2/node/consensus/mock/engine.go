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
	height  int64
	hash    types.Hash
	appHash types.Hash
}

type blkProp struct {
	height int64
	hash   types.Hash
	blk    *types.Block
	resCb  func(ack bool, appHash *types.Hash) error
}

type ackFrom struct {
	fromPubKey []byte
	res        types.AckRes
}

type blkResult struct {
	commit func(ctx context.Context, rollback bool) error

	appHash types.Hash
	txRes   []types.TxResult
	// other updates for next block...
}

type Engine struct {
	bki types.BlockStore
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

	// if leader commit message comes before we finish executing ce.proposed
	// (and setting ce.prepared), remember the given apphash
	earlyCommitAppHash *types.Hash

	// as leader, we collect acks for the proposed block. after ack threshold,
	// we commit and clear proposed/prepared for next block
	acks []ackFrom
}

var NumValidatorsFake = 3 // B, C, D

func New(bs types.BlockStore, mp types.MemPool) *Engine {
	height, hash, appHash := bs.Best()
	lc := blkCommit{
		height:  height,
		hash:    hash,
		appHash: appHash,
	}
	return &Engine{
		lastCommit: lc,
		bki:        bs,
		mp:         mp,
		mined:      make(chan *types.QualifiedBlock, 1),
		validators: make([][]byte, NumValidatorsFake),
		exec:       &dummyExecEngine{},
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

func (ce *Engine) BeLeader(leader bool) {
	ce.leader.Store(leader)
}

// CommitBlock reports a full block to commit. This would be used when:
//  1. retrieval of a block following in an announcement for a new+next block
//  2. iterative block retrieval in catch-up / sync
func (ce *Engine) CommitBlock(blk *types.Block, appHash types.Hash) error {
	ctx := context.TODO()

	height := blk.Header.Height
	if ce.proposed == nil { // block sync
		// execute and commit if it is next
		if height != ce.lastCommit.height+1 {
			return fmt.Errorf("block at height %d does not follow %d", height, ce.lastCommit.height)
		}

		hash := blk.Header.Hash()

		commitFn, appHash, _ /*res*/, err := ce.exec.ExecBlock(blk)
		if err != nil {
			return err
		}
		if err = commitFn(ctx, false); err != nil {
			return err
		}
		// todo store apphash and tx res

		ce.lastCommit = blkCommit{
			height:  blk.Header.Height,
			hash:    hash,
			appHash: appHash,
		}

		ce.confirmBlkTxns(blk)

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

	if ce.prepared != nil { // we were ready
		if ce.prepared.appHash != appHash {
			ce.rollbackPrepared(ctx)
			return fmt.Errorf("block %d appHash mismatch: leader %s != computed %s",
				height, appHash, ce.prepared.appHash)
		}
		ce.commitPrepared(ctx)
		return nil
	}

	// this is super dumb and why we need changes to go through loop with channels
	ce.earlyCommitAppHash = &appHash

	return nil
}

func (ce *Engine) confirmBlkTxns(blk *types.Block) {
	blkHash := blk.Header.Hash()
	height := blk.Header.Height

	log.Printf("confirming %d transactions in block %d (%v)", len(blk.Txns), height, blkHash)
	for _, txn := range blk.Txns {
		txHash := types.HashBytes(txn)
		ce.mp.Store(txHash, nil) // rm from mempool
	}

	// rawBlk := types.EncodeBlock(blk)
	ce.bki.Store(blk, fakeAppHash(height))
}
