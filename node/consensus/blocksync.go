package consensus

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kwilteam/kwil-db/node/types"
)

// Leader has to figure out the best height to ensure that it is
// caught up with the network, before it can start proposing blocks.
// Else it can propose a block at a height that is already finalized
// leading to a fork.
func (ce *ConsensusEngine) doBlockSync(ctx context.Context) error {
	if ce.role.Load() == types.RoleLeader {
		// TODO: which validator set we should use here? whatever we have in the state?
		// what if the current validators are not the same as the ones in the state?
		// and they don't respond to the status request?
		if len(ce.validatorSet) == 1 {
			return nil // we are the network
		}
		// figure out the best height to sync with the network
		// before starting to request blocks from the network.
		bestHeight, err := ce.discoverBestHeight(ctx)
		if err != nil {
			return fmt.Errorf("failed to discover the network's best height: %w", err)
		}

		if bestHeight <= ce.state.lc.height {
			// replay blocks from the network to catch up with the network.
			ce.log.Info("Leader is up to date with the network", "height", ce.state.lc.height)

			return nil
		}

		// replay blocks from the network to catch up with the network.
		return ce.syncBlocksUntilHeight(ctx, ce.state.lc.height+1, bestHeight)
	}

	// Validators and sentry nodes can do best effort block sync
	// and start the consensus engine at the latest height. If they
	// are behind, they can catch up as they process blocks.
	return ce.replayBlockFromNetwork(ctx)
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
				if ce.inSync.Load() {
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
	startHeight := ce.state.lc.height + 1
	t0 := time.Now()

	for {
		_, appHash, rawblk, err := ce.blkRequester(ctx, ce.state.lc.height+1)
		if err != nil { // all kinds of errors?
			ce.log.Info("Error requesting block from network", "height", ce.state.lc.height+1, "error", err)
			break // no more blocks to sync from network.
		}

		if ce.state.lc.height != 0 && appHash.IsZero() {
			return nil
		}

		blk, err := types.DecodeBlock(rawblk)
		if err != nil {
			return fmt.Errorf("failed to decode block: %w", err)
		}

		if err := ce.processAndCommit(ctx, blk, appHash); err != nil {
			return err
		}
	}

	ce.log.Info("Block sync completed", "startHeight", startHeight, "endHeight", ce.state.lc.height, "duration", time.Since(t0))
	return nil
}

// replayBlockFromNetwork requests the next blocks from the network and processes it
// until it catches up with its peers.
func (ce *ConsensusEngine) syncBlocksUntilHeight(ctx context.Context, startHeight, endHeight int64) error {

	height := startHeight
	t0 := time.Now()

	for height <= endHeight {
		_, appHash, rawblk, err := ce.blkRequester(ctx, height)
		if err != nil { // all kinds of errors?
			ce.log.Info("Error requesting block from network", "height", ce.state.lc.height+1, "error", err)
			return nil // no more blocks to sync from network.
		}

		if ce.state.lc.height != 0 && appHash.IsZero() {
			return nil
		}

		blk, err := types.DecodeBlock(rawblk)
		if err != nil {
			return fmt.Errorf("failed to decode block: %w", err)
		}

		if err := ce.processAndCommit(ctx, blk, appHash); err != nil {
			return err
		}

		height++
	}

	ce.log.Info("Block sync completed", "startHeight", startHeight, "endHeight", endHeight, "duration", time.Since(t0))

	return nil
}
