package consensus

import (
	"bytes"
	"context"
	"log"
	"p2p/node/types"
	"slices"
	"time"
)

const (
	blockTxCount    = 50              // for "mining"
	dummyTxSize     = 123_000         // for broadcast
	dummyTxInterval = 1 * time.Second // broadcast freq
)

func (ce *Engine) mine(ctx context.Context, interval time.Duration) {
	var height int64
	const N = blockTxCount
	for {
		if ce.mp.Size() < N || !ce.leader.Load() {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
			continue
		}

		// Reap txns from mempool for the block
		txids, txns := ce.mp.ReapN(N)

		var prevHash types.Hash    // TODO
		var prevAppHash types.Hash // TODO
		blk := types.NewBlock(0, height, prevHash, prevAppHash, time.Now(), txns)
		hash := blk.Header.Hash()

		log.Printf("built new block with %d transactions at height %d (%v)", len(txids), height, hash)

		// rawBlk := types.EncodeBlock(blk)
		// ce.bki.Store(hash, height, rawBlk)

		ce.mined <- &types.QualifiedBlock{
			Block: blk,
			Hash:  hash,
		}
		height++
	}
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
	af := ackFrom{fromPubKey, ack}
	if idx == -1 { // new!
		ce.acks = append(ce.acks, af)
	} else {
		log.Printf("replacing known ACK from %x: %v", fromPubKey, ack)
		ce.acks[idx] = af
	}

	// TODO: again, send to event loop, so this can trigger commit if threshold reached.
	ce.tallyAcks()
}

func (ce *Engine) tallyAcks() {
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

		if a.res.AppHash == wantAppHash {
			confirms++
		} else {
			log.Println("VALIDATOR DISAGREES ON APPHASH!", a.res.AppHash)
		}
	}

	if confirms >= ce.validatorThreshold() {
		// commit locally and broadcast commit message to validators
	}
}

func (ce *Engine) validatorThreshold() int {
	return len(ce.validators)/2 + 1
}
