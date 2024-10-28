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
	haltChan chan struct{} // can take a msg or reason for halting the network

	// interfaces
	mempool       Mempool
	blockStore    BlockStore
	blockExecutor BlockExecutor

	proposalBroadcaster ProposalBroadcaster
	blkAnnouncer        BlkAnnouncer
	ackBroadcaster      AckBroadcaster
	blkRequester        BlkRequester

	// logger log.Logger
}

// ProposalBroadcaster broadcasts the new block proposal message to the network
type ProposalBroadcaster func(ctx context.Context, blk *types.Block, id peer.ID)

// BlkAnnouncer broadcasts the new committed block to the network using the blockAnn message
type BlkAnnouncer func(ctx context.Context, blk *types.Block, appHash types.Hash, from peer.ID)

// AckBroadcaster gossips the ack/nack messages to the network
type AckBroadcaster func(ack bool, height int64, blkID types.Hash, appHash *types.Hash) error

// BlkRequester requests the block from the network based on the height
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
}

type blockResult struct {
	ack       bool
	appHash   types.Hash
	txResults []types.TxResult
}

type lastCommit struct {
	height  int64
	blkHash types.Hash

	appHash types.Hash
	blk     *types.Block // why is this needed? can be fetched from the blockstore too.
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
	}
}

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

	// start mining
	ce.startMining(ctx)

	// start the event loop
	ce.runEventLoop(ctx)
}

// runEventLoop starts the event loop for the consensus engine.
// Below are the event triggers that nodes can receive depending on their role:
// Leader:
//   - Acks
//
// Validator:
//   - BlockProp
//   - BlockAnn
//
// Sentry:
//   - BlockAnn
//
// Apart from the above events, the node also periodically checks if it needs to
// catchup with the network and reannounce the messages.
func (ce *ConsensusEngine) runEventLoop(ctx context.Context) error {
	// TODO: make these configurable?
	catchUpTicker := time.NewTicker(1 * time.Second)
	reannounceTicker := time.NewTicker(200 * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ce.haltChan:
			// Halt the network
			fmt.Println("Received halt signal, stopping the consensus engine")
			return nil

		case <-catchUpTicker.C:
			ce.doCatchup(ctx) // better name??

		case <-reannounceTicker.C:
			ce.reannounceMsgs(ctx)

		case m := <-ce.msgChan:
			ce.handleConsensusMessages(ctx, m)
		}
	}
}

// startMining starts the mining process based on the role of the node.
func (ce *ConsensusEngine) startMining(ctx context.Context) {
	// Start the mining process if the node is a leader
	// validators and sentry nodes get activated when they receive a block proposal or block announce msgs.
	if ce.role == types.RoleLeader {
		fmt.Println("Starting the leader node")
		go ce.startNewRound(ctx)
	} else {
		fmt.Println("Starting the validator/sentry node")
	}
}

// handleConsensusMessages handles the consensus messages based on the message type.
func (ce *ConsensusEngine) handleConsensusMessages(ctx context.Context, msg consensusMessage) {
	// validate the message
	// based on the message type, process the message
	fmt.Println("Consensus message received: ", msg.MsgType, msg.Sender)

	switch msg.MsgType {
	case "block_proposal":
		blkPropMsg, ok := msg.Msg.(*blockProposal)
		if !ok {
			fmt.Println("Invalid block proposal message")
			return // ignore the message
		}
		go ce.processBlockProposal(ctx, blkPropMsg) // This triggers the processing of the block proposal

	case "vote":
		// only leader should receive votes
		if ce.role != types.RoleLeader {
			return
		}

		vote, ok := msg.Msg.(*vote)
		if !ok {
			fmt.Println("Invalid vote message")
			return
		}

		if err := ce.addVote(ctx, vote, string(msg.Sender)); err != nil {
			fmt.Println("Error adding vote: ", vote, err)
			return
		}

	case "block_ann":
		blkAnn, ok := msg.Msg.(*blockAnnounce)
		if !ok {
			fmt.Println("Invalid block announce message")
			return
		}

		go ce.commitBlock(blkAnn.blk, blkAnn.appHash)
	}
}

// catchup syncs the node first with the local blockstore and then with the network.
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
	if err := ce.replayLocalBlocks(); err != nil {
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
		// WHat does it mean to have a dirty hash here? how would you handle it? roll back mechanics?
		ce.state.lc.appHash = state.AppHash
		ce.state.lc.blkHash = hash
	}

	return nil
}

// replayBlocks replays all the blocks from the blockstore if the app hasn't played all the blocks yet.
func (ce *ConsensusEngine) replayLocalBlocks() error {
	for {
		_, blk, appHash, err := ce.blockStore.GetByHeight(ce.state.lc.height + 1)
		if err != nil { // no more blocks to replay
			return nil
		}

		err = ce.processAndCommit(blk, appHash)
		if err != nil {
			return fmt.Errorf("failed replaying block: %v", err)
		}
	}
}

// replayBlockFromNetwork requests the next blocks from the network and processes it
// until it catches up with its peers.
func (ce *ConsensusEngine) replayBlockFromNetwork(ctx context.Context) error {
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
		if ce.state.blkProp != nil && ce.state.blockRes != nil &&
			ce.state.blockRes.appHash != types.ZeroHash &&
			ce.networkHeight <= ce.state.lc.height { // To ensure that we are not reannouncing the acks for very stale blocks
			// TODO: rethink what to broadcast here ack/nack, how do u remember the ack/nack
			fmt.Println("Reannouncing the acks", ce.state.blkProp.height, ce.state.blkProp.blkHash, ce.state.blockRes.appHash)
			go ce.ackBroadcaster(true, ce.state.blkProp.height, ce.state.blkProp.blkHash, &ce.state.blockRes.appHash)
		}
	}
}

func (ce *ConsensusEngine) doCatchup(ctx context.Context) {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.role != types.RoleLeader {
		if ce.state.blkProp == nil && ce.state.blockRes == nil {
			// catchup if needed with the leader/network.
			fmt.Println("Requesting block from network (staggered validator): ", ce.state.lc.height+1)
			ce.replayBlockFromNetwork(ctx)
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
