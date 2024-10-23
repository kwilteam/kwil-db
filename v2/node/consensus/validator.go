package consensus

import (
	"log"
	"p2p/node/types"
)

// AcceptProposalID decides if we should retrieve the content of a solicited
// block proposal. It's yes if we are not loader, we do not already have a
// proposal being processed, and the proposal builds on our best block.
func (ce *Engine) AcceptProposalID(height int64, prevHash types.Hash) bool {
	if ce.leader.Load() {
		return false // not for leader
	}

	ce.mtx.RLock()
	defer ce.mtx.RUnlock()
	if ce.proposed != nil {
		log.Println("block proposal already exists")
		return false
	}
	// initial block must precede genesis
	if height == 1 || prevHash.IsZero() {
		return ce.lastCommit.hash.IsZero()
	}

	if height != ce.lastCommit.height+1 {
		return false
	}
	return prevHash == ce.lastCommit.hash
}

// ProcessProposal handles a full block proposal from the leader. Only a
// validator should use this method, not leader or sentry. This validates the
// proposal, ensuring that it is for the next block (by height and previous
// block hash), and beings executing the block. When execution is complete, the
// res callback function is called with ACK+appHash/nACK, which in the context
// of the node will send the outcome back to the leader where validator
// responses are tallied.
func (ce *Engine) ProcessProposal(blk *types.Block, res func(ack bool, appHash types.Hash) error) {
	if ce.leader.Load() {
		return // not for leader
	}

	ce.mtx.Lock()
	defer ce.mtx.Unlock()

	if ce.proposed != nil {
		log.Println("block proposal already exists")
		return
	}
	if blk.Header.Height != ce.lastCommit.height+1 {
		log.Printf("proposal for height %d does not follow %d", blk.Header.Height, ce.lastCommit.height)
		return
	}

	blkHash := blk.Header.Hash()

	// we will then begin execution, and later report with ack/nack

	ce.proposed = &blkProp{
		height: blk.Header.Height,
		hash:   blkHash,
		blk:    blk,
		resCb:  res,
	}

	// OR

	// ce.evtChan <- *&blkProp{...}

	// ce event loop will send ACK+appHash or NACK.

	// ce should have some handle to p2p, like a function or channel into an
	// outgoing p2p msg loop.
}

// AcceptCommit is used for a validator to handle a committed block
// announcement. The result should indicate if if the block should be fetched.
// This will return false when ANY of the following are the case:
//
//  1. (validator) we had the proposed block, which we will commit when ready
//  2. this is not the next block in our local store
//
// This will return true if we should fetch the block, which is the case if BOTH
// we did not have a proposal for this block, and it is the next in our block store.
func (ce *Engine) AcceptCommit(height int64, blkHash types.Hash) (fetch bool) {
	if ce.leader.Load() {
		return // not for leader
	}

	if ce.proposed != nil && ce.proposed.hash == blkHash {
		// this should signal for CE to commit the block once it is executed.
		return false
	}
	if height != ce.lastCommit.height+1 {
		return false
	}
	return !ce.bki.Have(blkHash)
}
