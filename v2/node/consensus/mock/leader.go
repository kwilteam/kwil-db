package dummyce

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"math/rand/v2"
	"p2p/node/types"
	"slices"
	"time"
)

const (
	blockTxCount = 30 // for "mining"
)

func fakeAppHash(height int64) types.Hash {
	return types.HashBytes(binary.LittleEndian.AppendUint64(nil, uint64(height)))
}

func (ce *Engine) haveProposed() bool {
	ce.mtx.Lock()
	defer ce.mtx.Unlock()
	return ce.proposed != nil
}

func (ce *Engine) mine(ctx context.Context, interval time.Duration) {
	height := ce.lastCommit.height + 1
	const N = blockTxCount
	for {
		if ce.mp.Size() < N || !ce.leader.Load() || ce.haveProposed() {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
			continue
		}

		// Reap txns from mempool for the block
		txids, txns := ce.mp.ReapN(N)

		prevHash := ce.lastCommit.hash
		prevAppHash := fakeAppHash(ce.lastCommit.height)
		blk := types.NewBlock(height, prevHash, prevAppHash, time.Now().UTC(), txns)
		hash := blk.Header.Hash()

		log.Printf("built new block with %d transactions at height %d (%v)", len(txids), height, hash)

		// rawBlk := types.EncodeBlock(blk)
		// ce.bki.Store(hash, height, rawBlk)

		ce.proposed = &blkProp{
			height: blk.Header.Height,
			hash:   hash,
			blk:    blk,
			resCb:  func(ack bool, appHash *types.Hash) error { return nil },
		}

		ce.mined <- &types.QualifiedBlock{
			Block:    blk,
			Hash:     hash,
			Proposed: true,
		}
		height++

		// now "execute" and put result in prepared for later commit
		go func() {
			commitFn, appHash, txRes, err := ce.exec.ExecBlock(blk)
			if err != nil {
				log.Println("ExecBlock failed:", err)
			}

			ce.mtx.Lock()
			defer ce.mtx.Unlock()
			ce.prepared = &blkResult{commit: commitFn, txRes: txRes, appHash: appHash}
			ce.tallyAcksAndCommit() // and commit if validator threshold beat us
		}()
	}
}

type dummyExecEngine struct{}

func (ee *dummyExecEngine) ExecBlock(blk *types.Block) (commitFn func(context.Context, bool) error, appHash types.Hash, res []types.TxResult, err error) {
	time.Sleep(time.Duration(rand.Int64N(200)+10) * time.Millisecond)
	commitFn = func(context.Context, bool) error { return nil } // this would be dbTx.Commit()
	res = []types.TxResult{}
	appHash = types.HashBytes(binary.LittleEndian.AppendUint64(nil, uint64(blk.Header.Height)))
	return commitFn, appHash, res, nil
}

func (ce *Engine) BlockLeaderStream() <-chan *types.QualifiedBlock {
	return ce.mined
}

// ProcessACK is used for leader to register a validator's ack message (the
// async response to leader's block proposal).
func (ce *Engine) ProcessACK(fromPubKey []byte, ack types.AckRes) {
	if !ce.leader.Load() {
		return
	}

	ce.mtx.Lock()
	defer ce.mtx.Unlock()
	idx := slices.IndexFunc(ce.acks, func(r ackFrom) bool {
		return bytes.Equal(r.fromPubKey, fromPubKey)
	})
	if ack.ACK && ack.AppHash == nil {
		log.Println("ACK must specify AppHash")
		return
	}
	af := ackFrom{fromPubKey, ack}
	if idx == -1 { // new!
		log.Printf("new ACK received from %x: %v", fromPubKey, ack)
		ce.acks = append(ce.acks, af)
	} else {
		log.Printf("replacing known ACK from %x: %v", fromPubKey, ack)
		ce.acks[idx] = af
	}

	// TODO: again, send to event loop, so this can trigger commit if threshold reached.
	ce.tallyAcksAndCommit()
}

func (ce *Engine) tallyAcksAndCommit() {
	if ce.prepared == nil {
		return // validator beat us! commit when we finish or next reporting validator hits threshold
	}

	wantAppHash := ce.prepared.appHash

	var acks, nacks, confirms int
	for _, a := range ce.acks {
		if !a.res.ACK {
			nacks++
			continue
		}
		acks++

		if *a.res.AppHash == wantAppHash {
			confirms++
		} else {
			log.Println("VALIDATOR DISAGREES ON APPHASH!", wantAppHash, a.res.AppHash)
		}
	}

	log.Println("tally: acks", acks, "nacks", nacks, "confirms", confirms)

	if confirms >= ce.validatorThreshold() {
		// commit locally and broadcast commit message to validators
		log.Println("threshold ACKs received, committing...")
		prop := ce.proposed
		ce.commitPrepared(context.TODO())
		ce.mined <- &types.QualifiedBlock{
			Block:    prop.blk,
			Hash:     prop.hash,
			Proposed: false, // committed
			AppHash:  &wantAppHash,
		}
	}
}

func (ce *Engine) validatorThreshold() int {
	return len(ce.validators)/2 + 1
}

func (ce *Engine) commitPrepared(ctx context.Context) {
	const rollback = false
	if err := ce.prepared.commit(ctx, rollback); err != nil {
		log.Fatal("commit:", err) // we're screwed
		return
	}
	log.Println("db commit success!")

	// end of round!
	ce.lastCommit = blkCommit{
		height:  ce.proposed.height,
		hash:    ce.proposed.hash,
		appHash: ce.prepared.appHash,
	}

	ce.confirmBlkTxns(ce.proposed.blk)

	ce.acks = nil
	ce.proposed = nil
	ce.prepared = nil
	ce.earlyCommitAppHash = nil
}

func (ce *Engine) rollbackPrepared(ctx context.Context) {
	const rollback = true
	if err := ce.prepared.commit(ctx, rollback); err != nil {
		log.Fatal("rollback failed:", err) // we're screwed
		return
	}

	ce.acks = nil
	ce.proposed = nil
	ce.prepared = nil
	ce.earlyCommitAppHash = nil
}
