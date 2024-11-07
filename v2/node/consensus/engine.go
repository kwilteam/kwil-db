package consensus

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"kwil/crypto"
	"kwil/log"
	"kwil/node/types"
	ktypes "kwil/types"

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
	role   atomic.Value // types.Role, role can change over the lifetime of the node
	host   peer.ID
	dir    string
	signer crypto.PrivateKey
	leader crypto.PublicKey

	proposeTimeout time.Duration

	networkHeight atomic.Int64
	validatorSet  map[string]types.Validator

	// stores state machine state for the consensus engine
	state state

	// copy of the state info for the p2p layer usage.
	stateInfo StateInfo

	// Channels
	msgChan   chan consensusMessage
	haltChan  chan struct{} // can take a msg or reason for halting the network
	resetChan chan int64    // to reset the state of the consensus engine

	// interfaces
	log           log.Logger
	mempool       Mempool
	blockStore    BlockStore
	blockExecutor BlockExecutor

	proposalBroadcaster ProposalBroadcaster
	blkAnnouncer        BlkAnnouncer
	ackBroadcaster      AckBroadcaster
	blkRequester        BlkRequester
	rstStateBroadcaster ResetStateBroadcaster

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

type ResetStateBroadcaster func(height int64) error

type Status string

const (
	Proposed  Status = "proposed"  // SM has a proposed block for the current height
	Executed  Status = "executed"  // SM has executed the proposed block
	Committed Status = "committed" // SM has committed the block
)

// StateInfo contains the state information required by the p2p layer to
// download the blocks and notify the consensus engine about the incoming blocks.
type StateInfo struct {
	// mtx protects the below fields and should be locked by the consensus engine
	// only when updating the state and the locks should be released immediately.
	mtx sync.RWMutex

	// height of the last committed block
	height int64

	// status of the consensus engine
	status Status

	// proposed block for the current height
	blkProp *blockProposal
}

// Consensus state that is applicable for processing the blioc at a speociifc height.
type state struct {
	mtx sync.RWMutex

	cancelFunc context.CancelFunc
	// consensusTx sql.Tx

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
	txResults []ktypes.TxResult
}

type lastCommit struct {
	height  int64
	blkHash types.Hash

	appHash types.Hash
	blk     *types.Block // why is this needed? can be fetched from the blockstore too.
}

// Config is the struct given to the constructor, [New].
type Config struct {
	// Role is the role of the node in the consensus engine.
	Role types.Role
	// Signer is the private key of the node.
	Signer crypto.PrivateKey
	// HostID is the ID of the node.
	HostID peer.ID // TODO: not the libp2p2 peer.ID type
	// Dir is the directory where the node stores its data.
	Dir string
	// Leader is the public key of the leader.
	Leader crypto.PublicKey
	// Mempool is the mempool of the node.
	Mempool Mempool
	// BlockStore is the blockstore of the node.
	BlockStore BlockStore
	// ValidatorSet is the set of validators in the network.
	ValidatorSet map[string]types.Validator
	// Logger is the logger of the node.
	Logger log.Logger

	// ProposeTimeout is the timeout for proposing a block.
	ProposeTimeout time.Duration
}

// New creates a new consensus engine.
func New(cfg *Config) *ConsensusEngine {
	logger := cfg.Logger
	if logger == nil {
		// logger = log.DiscardLogger // for prod
		logger = log.New(log.WithName("CONS"), log.WithLevel(log.LevelDebug),
			log.WithWriter(os.Stdout), log.WithFormat(log.FormatUnstructured))
	}

	be := newBlockExecutor()
	// rethink how this state is initialized
	ce := &ConsensusEngine{
		host:           cfg.HostID,
		dir:            cfg.Dir,
		signer:         cfg.Signer,
		leader:         cfg.Leader,
		proposeTimeout: cfg.ProposeTimeout,
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
		stateInfo: StateInfo{
			height:  0,
			status:  Committed,
			blkProp: nil,
		},
		validatorSet: maps.Clone(cfg.ValidatorSet),
		msgChan:      make(chan consensusMessage, 1), // buffer size??
		haltChan:     make(chan struct{}, 1),
		resetChan:    make(chan int64, 1),
		// interfaces
		mempool:       cfg.Mempool,
		blockStore:    cfg.BlockStore,
		blockExecutor: be,
		log:           logger,
	}

	ce.role.Store(cfg.Role)
	ce.networkHeight.Store(0)

	return ce
}

func (ce *ConsensusEngine) Start(ctx context.Context, proposerBroadcaster ProposalBroadcaster,
	blkAnnouncer BlkAnnouncer, ackBroadcaster AckBroadcaster, blkRequester BlkRequester, stateResetter ResetStateBroadcaster) {
	ce.proposalBroadcaster = proposerBroadcaster
	ce.blkAnnouncer = blkAnnouncer
	ce.ackBroadcaster = ackBroadcaster
	ce.blkRequester = blkRequester
	ce.rstStateBroadcaster = stateResetter

	// Fast catchup the node with the network height
	if err := ce.catchup(ctx); err != nil {
		ce.log.Errorf("Error catching up: %v", err)
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
	catchUpTicker := time.NewTicker(5 * time.Second)
	reannounceTicker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-ctx.Done():
			ce.log.Info("Shutting down the consensus engine")
			return nil

		case <-ce.haltChan:
			// Halt the network
			ce.log.Error("Received halt signal, stopping the consensus engine")
			return nil

		case <-catchUpTicker.C:
			ce.doCatchup(ctx) // better name??

		case <-reannounceTicker.C:
			ce.reannounceMsgs(ctx)

		case height := <-ce.resetChan:
			ce.resetBlockProp(height)

		case m := <-ce.msgChan:
			ce.handleConsensusMessages(ctx, m)
		}
	}
}

// startMining starts the mining process based on the role of the node.
func (ce *ConsensusEngine) startMining(ctx context.Context) {
	// Start the mining process if the node is a leader
	// validators and sentry nodes get activated when they receive a block proposal or block announce msgs.
	if ce.role.Load() == types.RoleLeader {
		ce.log.Infof("Starting the leader node")
		go ce.startNewRound(ctx)
	} else {
		ce.log.Infof("Starting the validator/sentry node")
	}
}

// handleConsensusMessages handles the consensus messages based on the message type.
func (ce *ConsensusEngine) handleConsensusMessages(ctx context.Context, msg consensusMessage) {
	// validate the message
	// based on the message type, process the message
	ce.log.Info("Consensus message received", "type", msg.MsgType, "sender", hex.EncodeToString(msg.Sender))

	switch msg.MsgType {
	case "block_proposal":
		blkPropMsg, ok := msg.Msg.(*blockProposal)
		if !ok {
			ce.log.Warnf("Invalid block proposal message")
			return // ignore the message
		}
		// go ce.processBlockProposal(ctx, blkPropMsg) // This triggers the processing of the block proposal
		ce.processBlockProposal(ctx, blkPropMsg)

	case "vote":
		// only leader should receive votes
		if ce.role.Load() != types.RoleLeader {
			return
		}

		vote, ok := msg.Msg.(*vote)
		if !ok {
			ce.log.Warnf("Invalid vote message")
			return
		}

		if err := ce.addVote(ctx, vote, hex.EncodeToString(msg.Sender)); err != nil {
			ce.log.Error("Error adding vote", "vote", vote, "error", err)
			return
		}

	case "block_ann":
		blkAnn, ok := msg.Msg.(*blockAnnounce)
		if !ok {
			ce.log.Error("Invalid block announce message")
			return
		}

		if err := ce.commitBlock(blkAnn.blk, blkAnn.appHash); err != nil {
			ce.log.Error("Error processing committing block", "error", err)
			return
		}
	}

}

// resetBlockProp aborts the block execution and resets the state to the last committed block.
func (ce *ConsensusEngine) resetBlockProp(height int64) {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// If we are currently executing any transactions corresponding to the blk at height +1
	// 1. Cancel the execution context -> so that the transactions stop
	// 2. Rollback the consensus tx
	// 3. Reset the blkProp and blockRes
	// 4. This should never happen after the commit phase, (blk should have never made it to the blockstore)

	if ce.state.lc.height == height {
		if ce.state.blkProp != nil {
			// first cancel the context
			ce.state.cancelFunc()
			// rollback the pg tx
			// ce.state.consensusTx.Rollback()

			// reset the blkProp and blockRes
			ce.log.Info("Resetting the block proposal", "height", height)
			ce.resetState()
			// no need to update the last commit info as commit phase is not reached yet
		}
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

	ce.log.Info("Initial APP State", "height", ce.state.appState.Height, "appHash", ce.state.appState.AppHash)
	// Replay blocks from the blockstore.
	startHeight := ce.state.lc.height + 1
	t0 := time.Now()

	if err := ce.replayLocalBlocks(); err != nil {
		return err
	}
	ce.log.Info("Replayed blocks from the blockstore", "from", startHeight, "to (excluding)", ce.state.lc.height+1, "elapsed", time.Since(t0), "appHash", ce.state.lc.appHash)

	startHeight = ce.state.lc.height + 1
	t0 = time.Now()
	// Replay blocks from the network
	if err := ce.replayBlockFromNetwork(ctx); err != nil {
		return err
	}
	ce.log.Info("Replayed blocks from the network", "from", startHeight, "to (excluding)", ce.state.lc.height+1, "elapsed", time.Since(t0), "appHash", ce.state.lc.appHash)

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
		ce.state.lc.height = state.Height

		// TODO: check for any uncommitted transaction in PG and rollback
		// retrieve the block from the blockstore
		hash, blk, _, err := ce.blockStore.GetByHeight(ce.state.lc.height)
		if err != nil {
			// This is not possible. The state.json will have the height, only when the block is committed.
			return err
		}

		if state.AppHash == dirtyHash {
			// This implies that the commit was successful, but the final appHash is not yet updated.
			state.AppHash = hash
			ce.persistAppState()
		}

		ce.state.lc.blk = blk
		ce.state.lc.appHash = state.AppHash
		ce.state.lc.blkHash = hash

		ce.stateInfo.height = state.Height
		ce.stateInfo.status = Committed
		ce.stateInfo.blkProp = nil
	}

	return nil
}

// replayBlocks replays all the blocks from the blockstore if the app hasn't played all the blocks yet.
func (ce *ConsensusEngine) replayLocalBlocks() error {
	for {
		_, blk, appHash, err := ce.blockStore.GetByHeight(ce.state.lc.height + 1)
		if err != nil {
			if !errors.Is(err, types.ErrNotFound) {
				return fmt.Errorf("unexpected blockstore error: %w", err)
			}
			return nil // no more blocks to replay
		}

		err = ce.processAndCommit(blk, appHash)
		if err != nil {
			return fmt.Errorf("failed replaying block: %w", err)
		}
	}
}

// replayBlockFromNetwork requests the next blocks from the network and processes it
// until it catches up with its peers.
func (ce *ConsensusEngine) replayBlockFromNetwork(ctx context.Context) error {
	for {
		_, appHash, rawblk, err := ce.blkRequester(ctx, ce.state.lc.height+1)
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

	if ce.role.Load() == types.RoleLeader {
		// reannounce the blkProp message if the node is still waiting for the votes
		if ce.state.blkProp != nil {
			go ce.proposalBroadcaster(ctx, ce.state.blkProp.blk, ce.host)
		}
		if ce.state.lc.height > 0 {
			// Announce block commit message for the last committed block
			go ce.blkAnnouncer(ctx, ce.state.lc.blk, ce.state.lc.appHash, ce.host) // TODO: can be made infrequent
		}
		return
	}

	if ce.role.Load() == types.RoleValidator {
		// reannounce the acks, if still waiting for the commit message
		if ce.state.blkProp != nil && ce.state.blockRes != nil &&
			ce.state.blockRes.appHash != types.ZeroHash &&
			ce.networkHeight.Load() <= ce.state.lc.height { // To ensure that we are not reannouncing the acks for very stale blocks
			// TODO: rethink what to broadcast here ack/nack, how do u remember the ack/nack
			ce.log.Info("Reannouncing ACKs", "height", ce.state.blkProp.height, "hash", ce.state.blkProp.blkHash,
				"appHash", ce.state.blockRes.appHash)
			go ce.ackBroadcaster(true, ce.state.blkProp.height, ce.state.blkProp.blkHash, &ce.state.blockRes.appHash)
		}
	}
}

func (ce *ConsensusEngine) doCatchup(ctx context.Context) {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.state.lc.height >= ce.networkHeight.Load() {
		return
	}

	if ce.role.Load() != types.RoleLeader {
		if ce.state.blkProp == nil && ce.state.blockRes == nil {
			// catchup if needed with the leader/network.
			startHeight := ce.state.lc.height + 1
			t0 := time.Now()
			ce.replayBlockFromNetwork(ctx)
			ce.log.Info("Downloaded blocks from the network", "from", startHeight, "to (excluding)", ce.state.lc.height+1, "time", time.Since(t0), "appHash", ce.state.lc.appHash)
		}
	}
}

func (ce *ConsensusEngine) updateNetworkHeight(height int64) {
	if height > ce.networkHeight.Load() {
		ce.networkHeight.Store(height)
	}
}

func (ce *ConsensusEngine) requiredThreshold() int64 {
	return int64(len(ce.validatorSet)/2 + 1)
}
