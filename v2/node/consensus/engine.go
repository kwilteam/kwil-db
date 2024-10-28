package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"p2p/node/types"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// Use these accordingly
	MaxBlockSize = 4 * 1024 * 1024 // 1 MB
	blockTxCount = 50
)

// There are three phases in the consensus engine:
// 1. BlockProposalPhase:
//   - Depending on the role, the node is either preparing the block or waiting to receive the proposed block.
//
// 2. BlockExecutionPhase:
//   - Nodes enter this phase once they have the block to be processed. In this phase, all the transactions in the block are executed, and the appHash is generated. Leader then waits for the votes from the validators and validators respond with ack/nack.
//
// 3. BlockCommitPhase:
// - Once the leader receives the threshold acks with the same appHash as the leader, the block is committed and the leader broadcasts the blockAnn message to the network. Nodes that receive this message will enter into the commit phase where they verify the appHash and commit the block.
type ConsensusEngine struct {
	role          types.Role
	host          peer.ID
	dir           string
	pubKey        []byte
	networkHeight int64
	validatorSet  map[string]types.Validator

	// store last commit info
	state state

	// Channels
	msgChan  chan consensusMessage
	haltChan chan struct{}

	// interfaces
	mempool       Mempool
	blockStore    BlockStore
	blockExecutor BlockExecutor
	// indexer       Indexer

	// Broadcasters and Requesters
	proposalBroadcaster ProposalBroadcaster
	blkAnnouncer        BlkAnnouncer
	ackBroadcaster      AckBroadcaster
	blkRequester        BlkRequester
}

type ProposalBroadcaster func(ctx context.Context, blk *types.Block, id peer.ID)
type BlkAnnouncer func(ctx context.Context, blk *types.Block, appHash types.Hash, from peer.ID)
type AckBroadcaster func(ack bool, height int64, blkID types.Hash, appHash *types.Hash) error
type BlkRequester func(ctx context.Context, height int64) (types.Hash, types.Hash, []byte, error)

// Consensus state that is applicable for processing the blioc at a speociifc height.
type state struct {
	mtx sync.RWMutex

	blkProp  *blockProposal
	blockRes *blockResult
	lc       *lastCommit
	appState *appState

	// Votes: Applicable only to the leader
	// These are the Acks received from the validators.
	votes map[string]*vote
	// Move votes from the votes map to processedVotes map once the votes are processed.
	processedVotes map[string]*vote
}

type blockResult struct {
	appHash   types.Hash
	txResults []txResult
}

type lastCommit struct {
	height  int64
	blkHash types.Hash

	appHash types.Hash
	blk     *types.Block // why is this needed? can be fetched from the blockstore too.
}

type txResult struct {
	code uint16
	log  string
}

func New(role types.Role, hostID peer.ID, dir string, mempool Mempool, bs BlockStore,
	//indexer Indexer,
	vs map[string]types.Validator) *ConsensusEngine {

	pubKey, err := hostID.ExtractPublicKey()
	if err != nil {
		fmt.Println("Error extracting public key: ", err)
		return nil
	}
	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		fmt.Println("Error extracting public key bytes: ", err)
		return nil
	}

	// rethink how this state is initialized
	return &ConsensusEngine{
		role:   role,
		host:   hostID,
		dir:    dir,
		pubKey: pubKeyBytes,
		state: state{
			blkProp:  nil,
			blockRes: nil,
			lc: &lastCommit{
				height:  0,
				blkHash: types.ZeroHash,
				appHash: types.ZeroHash,
			},
			votes: make(map[string]*vote),
		},
		networkHeight: 0,
		validatorSet:  vs,
		msgChan:       make(chan consensusMessage, 1), // buffer size??
		haltChan:      make(chan struct{}, 1),
		// interfaces
		mempool:    mempool,
		blockStore: bs,
		// indexer:    indexer,
	}
}

// Start triggers the event loop for the consensus engine.
// Below are the event triggers based on the node's role:
// Leader:
//   - Acks
//
// Validator:
//   - BlockProp
//   - BlockAnn
//
// Sentry:
//   - BlockAnn
func (ce *ConsensusEngine) Start(ctx context.Context, proposerBroadcaster ProposalBroadcaster,
	blkAnnouncer BlkAnnouncer, ackBroadcaster AckBroadcaster, blkRequester BlkRequester) {
	ce.proposalBroadcaster = proposerBroadcaster
	ce.blkAnnouncer = blkAnnouncer
	ce.ackBroadcaster = ackBroadcaster
	ce.blkRequester = blkRequester

	// Fast catchup the node with the network height
	if err := ce.catchup(ctx); err != nil {
		fmt.Println("Error catching up: ", err)
		return
	}

	// Start the mining process if the node is a leader
	// validators and sentry nodes get activated when they receive a block proposal or block announce msgs.
	if ce.role == types.RoleLeader {
		fmt.Println("Starting the leader node")
		// start new round after 1 second
		time.Sleep(1 * time.Second)
		go ce.startNewRound(ctx)
	} else {
		fmt.Println("Starting the validator/sentry node")
	}

	// start the event loop
	var delay = 1 * time.Second
	for {
		select {
		case <-ce.haltChan:
			// Halt the network
			fmt.Println("Received halt signal, stopping the consensus engine")
			return
		case <-ctx.Done():
			return

		case <-time.After(delay):
			ce.reannounceMsgs(ctx)

		case m := <-ce.msgChan:
			fmt.Println("Msg received: ", m.MsgType, m.Sender)
			switch m.MsgType {
			case "block_proposal":
				blkPropMsg, ok := m.Msg.(*blockProposal)
				if !ok {
					fmt.Println("Block proposal message is not valid")
					continue
				}
				go ce.processBlockProposal(ctx, blkPropMsg) // This triggers the prepare phase

			case "vote":
				// only leader should receive votes
				if ce.role != types.RoleLeader {
					continue
				}

				vote, ok := m.Msg.(*vote)
				if !ok {
					continue
				}

				if err := ce.addVote(vote, string(m.Sender)); err != nil {
					fmt.Println("Error adding vote: ", vote, err)
					continue
				}

			case "block_ann":
				blkAnn, ok := m.Msg.(*blockAnnounce)
				if !ok {
					return
				}

				go ce.commitBlock(blkAnn.blk, blkAnn.appHash)
			}
		}
	}
}

func (ce *ConsensusEngine) updateNetworkHeight(height int64) {
	if height > ce.networkHeight {
		ce.networkHeight = height
	}
}

func (ce *ConsensusEngine) requiredThreshold() int64 {
	// TODO: update it
	// return int64(len(ce.validatorSet)/2 + 1)
	return 2
}

func (ce *ConsensusEngine) catchup(ctx context.Context) error {
	// Figure out the app state and initialize the node state.
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// initialize the consensus engine state
	if err := ce.init(); err != nil {
		return err
	}

	fmt.Println("Initial APP State: ", ce.state.appState.Height, ce.state.appState.AppHash.String())
	// Replay blocks from the blockstore.
	if err := ce.replayBlocks(); err != nil {
		return err
	}
	fmt.Println("Replayed blocks from the blockstore: ", ce.state.lc.height, ce.state.lc.appHash.String())

	// Replay blocks from the network
	if err := ce.replayBlockFromNetwork(ctx); err != nil {
		return err
	}
	fmt.Println("Replayed blocks from the network: ", ce.state.lc.height, ce.state.lc.appHash.String())

	return nil
}

// init initializes the node state based on the appState info.
func (ce *ConsensusEngine) init() error {
	state, err := ce.loadAppState()
	if err != nil {
		return err
	}

	ce.state.appState = state
	ce.persistAppState()

	// Retrieve the last commit info from the blockstore based on the appState info.
	if state.Height > 0 {
		if state.AppHash == dirtyHash {
			ce.state.lc.height = state.Height - 1
		} else {
			ce.state.lc.height = state.Height
		}

		// retrieve the block from the blockstore
		hash, blk, _, err := ce.blockStore.GetByHeight(ce.state.lc.height)
		if err != nil {
			return err
		}

		ce.state.lc.blk = blk
		ce.state.lc.appHash = state.AppHash
		ce.state.lc.blkHash = hash
	}

	return nil
}

// replayBlocks replays all the blocks from the blockstore if the app hasn't played all the blocks yet.
func (ce *ConsensusEngine) replayBlocks() error {
	_, blk, _, err := ce.blockStore.GetByHeight(ce.state.lc.height + 1)
	if err != nil {
		// no more blocks to replay
		return nil
	}

	for {
		// We need to fetch the next block for apphash validation
		// TODO: Ummm, but what if this is a leader trying to recover from crash, so trying to replay the blocks
		_, nextBlk, _, err := ce.blockStore.GetByHeight(ce.state.lc.height + 2)
		if err != nil {
			// App hash for the blk will be instead fetched from the network peers
			return nil
		}

		err = ce.processAndCommit(blk, nextBlk.Header.PrevAppHash)
		if err != nil {
			return fmt.Errorf("failed replaying block: %v", err)
		}

		blk = nextBlk
	}
}

// TODO: is this applicable only for the leader?
func (ce *ConsensusEngine) replayBlockFromNetwork(ctx context.Context) error {
	// Get the network's best height and replay the blocks from the network
	// or just keep asking for your next block from the network.
	for {
		_, appHash, rawblk, err := ce.blkRequester(ctx, ce.state.lc.height+1)
		fmt.Println("Requested block from network: ", ce.state.lc.height+1, appHash.String())
		if err != nil {
			return nil // no more blocks to sync from network.
		}

		if ce.state.lc.height != 0 && appHash == types.ZeroHash {
			return nil
		}

		blk, err := types.DecodeBlock(rawblk)
		if err != nil {
			return fmt.Errorf("failed to decode block: %v", err)
		}
		if err := ce.processAndCommit(blk, appHash); err != nil {
			return err
		}
	}
}

// TODO: probably different tickers for each kind of consensus message.
// Blocksync need to be way quicker, whereas the others need not be that frequent.
func (ce *ConsensusEngine) reannounceMsgs(ctx context.Context) {
	// Leader should reannounce the blkProp and blkAnn messages
	// Validators should reannounce the Ack messages
	ce.state.mtx.RLock()
	defer ce.state.mtx.RUnlock()

	if ce.role == types.RoleLeader {
		// reannounce the blkProp message if the node is still waiting for the votes
		if ce.state.blkProp != nil {
			go ce.proposalBroadcaster(ctx, ce.state.blkProp.blk, ce.host)
		}
		if ce.state.lc.height > 0 {
			// Announce block commit message for the last committed block
			go ce.blkAnnouncer(ctx, ce.state.lc.blk, ce.state.lc.appHash, ce.host)
		}
		return
	}

	if ce.role == types.RoleValidator {
		// reannounce the acks, if still waiting for the commit message
		if ce.state.blkProp != nil && ce.state.blockRes != nil && ce.state.blockRes.appHash != types.ZeroHash && ce.networkHeight <= ce.state.lc.height {
			// TODO: rethink what to broadcast here ack/nack, how do u remember the ack/nack
			fmt.Println("Reannouncing the acks", ce.state.blkProp.height, ce.state.blkProp.blkHash, ce.state.blockRes.appHash)
			go ce.ackBroadcaster(true, ce.state.blkProp.height, ce.state.blkProp.blkHash, &ce.state.blockRes.appHash)
		}
	}

	if ce.role != types.RoleLeader {
		if ce.state.blkProp == nil && ce.state.blockRes == nil {
			// catchup if needed with the leader/network.
			fmt.Println("Requesting block from network (staggered validator): ", ce.state.lc.height+1)
			ce.replayBlockFromNetwork(ctx)
		}
	}
}
