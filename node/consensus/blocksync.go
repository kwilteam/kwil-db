package consensus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jpillora/backoff"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// The leader must determine the best height to ensure it is synchronized with the network
// before proposing blocks. Otherwise, it might propose a block at an already finalized height,
// causing a fork. Two mechanisms are used to prevent the leader from committing a block at an
// already finalized height:
//  1. Manual checkpoints: The leader can be configured with a checkpoint height, which is the
//     height the leader must sync to before proposing blocks. The node operator should set the
//     checkpoint height based on the network's best height.
//  2. Validator feedback: If the leader proposes a block at an already finalized height, validators
//     will respond with a NACK, indicating the leader is out-of-sync. They can provide an OutOfSyncProof
//     as evidence, including the BlockHeader and the leader's signature of the block the validator is on.
//     This forces the leader to catch up to the validator's block height before proposing blocks again.
func (ce *ConsensusEngine) doBlockSync(ctx context.Context) error {
	if ce.role.Load() == types.RoleLeader {
		if err := ce.leaderBlockSync(ctx); err != nil {
			// not checking for BlockNotAvailable error here, as the leader
			// has to sync till the checkpoint height mandatorily. If blocks are
			// not available, leader should retry restarting node after some time.
			return fmt.Errorf("leader block sync failed: %w", err)
		}

		// verify the checkpoint hash and height
		if err := ce.VerifyCheckpoint(); err != nil {
			return fmt.Errorf("checkpoint verification failed: %w", err)
		}

		// Leader has synchronized with the network up to the checkpoint height,
		// now perform a best effort block sync to catch up further if needed.
	}

	// Validators and sentry nodes should perform a best effort block sync
	// to start the consensus engine at the most recent height. If they
	// are lagging, they can catch up while processing blocks.
	return ce.replayBlockFromNetwork(ctx, ce.syncBlock)
}

func (ce *ConsensusEngine) VerifyCheckpoint() error {
	// verify the checkpoint hash and height. If the checkpoints
	// are not set, return
	if ce.checkpoint.Height == 0 {
		return nil
	}

	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()

	height, hash := ce.stateInfo.lastCommit.height, ce.stateInfo.lastCommit.blkHash

	if height < ce.checkpoint.Height {
		return fmt.Errorf("checkpoint verification failed: height: %d [expected: %d]", height, ce.checkpoint.Height)
	}

	if height == ce.checkpoint.Height && hash != ce.checkpoint.Hash {
		return fmt.Errorf("checkpoint verification failed: height: %d hash: [expected: %s, curr: %s]", ce.checkpoint.Height, ce.checkpoint.Hash, hash)
	}

	return nil
}

func (ce *ConsensusEngine) leaderBlockSync(ctx context.Context) error {
	startHeight := ce.lastCommitHeight()
	ce.log.Info("Starting block sync", "height", startHeight+1)

	checkpoint := ce.checkpoint.Height
	if checkpoint <= startHeight {
		ce.log.Info("Leader is synced to the checkpoint", "height", startHeight)
		return nil
	}

	// Sync the leader to the checkpoint height.
	return ce.syncBlocksUntilHeight(ctx, startHeight+1, checkpoint)
}

// replayBlockFromNetwork attempts to synchronize the local node with the network by fetching
// and processing blocks from peers.
func (ce *ConsensusEngine) replayBlockFromNetwork(ctx context.Context, requester func(context.Context, int64) error) error {
	var startHeight, height int64
	startHeight = ce.lastCommitHeight() + 1
	height = startHeight
	t0 := time.Now()

	for {
		if err := requester(ctx, height); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			if errors.Is(err, types.ErrBlkNotFound) || errors.Is(err, types.ErrNotFound) || errors.Is(err, types.ErrPeersNotFound) {
				break // no peers have this block, assume block sync is complete, continue with consensus
			}
			ce.log.Warn("unexpected error requesting block from the network", "height", height, "error", err)
			return err
		}
		height++
	}

	ce.log.Info("Block sync completed", "startHeight", startHeight, "endHeight", height, "duration", time.Since(t0))
	return nil
}

// syncBlocksUntilHeight fetches and processes blocks from startHeight to endHeight,
// retrying if necessary until successful or the maximum retries are reached.
func (ce *ConsensusEngine) syncBlocksUntilHeight(ctx context.Context, startHeight, endHeight int64) error {
	height := startHeight
	t0 := time.Now()

	for height <= endHeight {
		if err := ce.syncBlockWithRetry(ctx, height); err != nil {
			return err
		}
		height++
	}

	ce.log.Info("Block sync completed", "startHeight", startHeight, "endHeight", endHeight, "duration", time.Since(t0))

	return nil
}

// syncBlockWithRetry fetches the specified block from the network and keeps retrying until
// the block is successfully retrieved from the network.
func (ce *ConsensusEngine) syncBlockWithRetry(ctx context.Context, height int64) error {
	blkID, rawBlk, ci, err := ce.getBlock(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to get block from the network: %w", err)
	}

	return ce.applyBlock(ctx, rawBlk, ci, blkID)
}

// syncBlock fetches the specified block from the network
func (ce *ConsensusEngine) syncBlock(ctx context.Context, height int64) error {
	blkID, rawblk, ci, _, err := ce.blkRequester(ctx, height)
	if err != nil {
		return fmt.Errorf("failed to get block from the network: %w", err)
	}

	return ce.applyBlock(ctx, rawblk, ci, blkID)
}

func (ce *ConsensusEngine) applyBlock(ctx context.Context, rawBlk []byte, ci *types.CommitInfo, blkID types.Hash) error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	blk, err := ktypes.DecodeBlock(rawBlk)
	if err != nil {
		return fmt.Errorf("failed to decode block: %w", err)
	}

	if err := ce.processAndCommit(ctx, blk, ci, blkID); err != nil {
		return fmt.Errorf("failed to apply block: %w", err)
	}

	return nil
}

func (ce *ConsensusEngine) getBlock(ctx context.Context, height int64) (blkID types.Hash, rawBlk []byte, ci *types.CommitInfo, err error) {
	err = blkRetrier(ctx, 15, func() error {
		blkID, rawBlk, ci, _, err = ce.blkRequester(ctx, height)
		return err
	})

	return blkID, rawBlk, ci, err
}

// retry will retry the function until it is successful, or reached the max retries
func blkRetrier(ctx context.Context, maxRetries int64, fn func() error) error {
	retrier := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    500 * time.Millisecond,
		Factor: 2,
		Jitter: true,
	}

	for {
		err := fn()
		if err == nil {
			return nil
		}

		if errors.Is(err, types.ErrBlkNotFound) {
			return err
		}

		// fail after maxRetries retries
		if retrier.Attempt() > float64(maxRetries) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retrier.Duration()):
		}
	}
}

/*
func (ce *ConsensusEngine) leaderBlockSync(ctx context.Context) error {
	if len(ce.validatorSet) == 1 {
		return nil // we are the network
	}

	startHeight := ce.lastCommitHeight()
	ce.log.Info("Starting block sync", "height", startHeight+1)

	// figure out the best height to sync with the network
	// before starting to request blocks from the network.
	bestHeight, err := ce.discoverBestHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover the network's best height: %w", err)
	}

	if bestHeight <= startHeight {
		// replay blocks from the network to catch up with the network.
		ce.log.Info("Leader is up to date with the network", "height", startHeight)
		return nil
	}

	return ce.syncBlocksUntilHeight(ctx, startHeight+1, bestHeight)
}

discoverBestHeight is a discovery process that leader uses to figure out the
latest network height from the validators.
func (ce *ConsensusEngine) discoverBestHeight(ctx context.Context) (int64, error) {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Broadcast a status message to the network, to which validators would respond with their status.
	go func() {
		ce.discoveryReqBroadcaster()

		delay := 500 * time.Millisecond
		for {
			select {
			case <-cancelCtx.Done():
				return
			case <-time.After(delay):
				if ce.InCatchup() { // if we are still in catchup, broadcast again
					ce.discoveryReqBroadcaster()
					delay = min(2*delay, 16*time.Second)
				}
			}
		}

	}()

	// Wait for the validators to respond with their best heights.
	// The leader would then sync up to the best height.

	heights := make(map[string]int64)
	var bestHeight int64 = 0

	for {
		select {
		case <-cancelCtx.Done():
			return -1, cancelCtx.Err()
		case msg := <-ce.bestHeightCh:
			// check if the msg is from a validator
			if _, ok := ce.validatorSet[hex.EncodeToString(msg.Sender)]; !ok {
				continue
			}

			sender := hex.EncodeToString(msg.Sender)
			heights[sender] = msg.BestHeight
			if msg.BestHeight > bestHeight {
				bestHeight = msg.BestHeight
			}

			if ce.hasMajorityFloor(len(heights)) {
				// found the best height
				// TODO: uniform logging of PubKey or IDs (hex.EncodeToString(msg.Sender) not same as peer.ID)
				ce.log.Info("Found the best height", "height", bestHeight, "fromPeer", sender)
				return bestHeight, nil
			}
		}
	}
}

*/
