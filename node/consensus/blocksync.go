package consensus

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jpillora/backoff"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

// Leader has to figure out the best height to ensure that it is
// caught up with the network, before it can start proposing blocks.
// Else it can propose a block at a height that is already finalized
// leading to a fork.
func (ce *ConsensusEngine) doBlockSync(ctx context.Context) error {
	if ce.role.Load() == types.RoleLeader {
		return ce.leaderBlockSync(ctx)
	}

	// Validators and sentry nodes can do best effort block sync
	// and start the consensus engine at the latest height. If they
	// are behind, they can catch up as they process blocks.
	return ce.replayBlockFromNetwork(ctx)
}

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

// replayBlockFromNetwork requests the next blocks from the network and processes it
// until it catches up with its peers.
func (ce *ConsensusEngine) replayBlockFromNetwork(ctx context.Context) error {
	var startHeight, height int64
	startHeight = ce.lastCommitHeight() + 1
	height = startHeight
	t0 := time.Now()

	for {
		if err := ce.syncBlock(ctx, height); err != nil {
			ce.log.Info("block request from the network failed", "height", height, "error", err)
			break
		}
		height++
	}

	ce.log.Info("Block sync completed", "startHeight", startHeight, "endHeight", height, "duration", time.Since(t0))
	return nil
}

// replayBlockFromNetwork requests the next blocks from the network and processes it
// until it catches up with its peers.
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

func (ce *ConsensusEngine) syncBlock(ctx context.Context, height int64) error {
	ce.log.Info("Requesting block from network (replay mode)", "height", height)

	_, rawblk, ci, err := ce.blkRequester(ctx, height)
	if err != nil {
		return err
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.lc.height != 0 && ci != nil && ci.AppHash.IsZero() {
		return nil
	}

	blk, err := ktypes.DecodeBlock(rawblk)
	if err != nil {
		return fmt.Errorf("failed to decode block: %w", err)
	}

	if err := ce.processAndCommit(ctx, blk, ci); err != nil {
		return err
	}

	return nil
}

// syncBlockWithRetry requests the network for the specified block and retries until the block
// is received and applied to the db.
func (ce *ConsensusEngine) syncBlockWithRetry(ctx context.Context, height int64) error {
	_, rawblk, ci, err := ce.getBlock(ctx, height)
	if err != nil { // all kinds of errors?
		ce.log.Info("Error requesting block from network", "height", height, "error", err)
		return fmt.Errorf("error requesting block from network: height : %d, error: %w", height, err)
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.lc.height != 0 && ci != nil && ci.AppHash.IsZero() {
		return nil
	}

	blk, err := ktypes.DecodeBlock(rawblk)
	if err != nil {
		return fmt.Errorf("failed to decode block: %w", err)
	}

	if err := ce.processAndCommit(ctx, blk, ci); err != nil {
		return err
	}

	return nil
}

func (ce *ConsensusEngine) getBlock(ctx context.Context, height int64) (blkID types.Hash, rawBlk []byte, ci *ktypes.CommitInfo, err error) {
	err = retry(ctx, 15, func() error {
		blkID, rawBlk, ci, err = ce.blkRequester(ctx, height)
		return err
	})

	return blkID, rawBlk, ci, err
}

// retry will retry the function until it is successful, or reached the max retries
func retry(ctx context.Context, maxRetries int64, fn func() error) error {
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
