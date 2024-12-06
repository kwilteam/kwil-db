package consensus

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

const (
	// Use these accordingly
	MaxBlockSize = 4 * 1024 * 1024 // 1 MB
	blockTxCount = 50
)

var zeroHash = types.Hash{}

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
	role    atomic.Value // types.Role, role can change over the lifetime of the node
	signer  auth.Signer
	privKey crypto.PrivateKey
	pubKey  crypto.PublicKey
	leader  crypto.PublicKey
	log     log.Logger

	proposeTimeout time.Duration

	networkHeight  atomic.Int64
	validatorSet   map[string]ktypes.Validator
	genesisAppHash types.Hash

	// stores state machine state for the consensus engine
	state  state
	inSync atomic.Bool // set when the node is still catching up with the network during bootstrapping

	// copy of the state info for the p2p layer usage.
	stateInfo StateInfo

	chainCtx *common.ChainContext

	// Channels
	msgChan      chan consensusMessage
	haltChan     chan struct{}      // can take a msg or reason for halting the network
	resetChan    chan int64         // to reset the state of the consensus engine
	bestHeightCh chan *discoveryMsg // to sync the leader with the network

	// interfaces
	db          DB
	mempool     Mempool
	blockStore  BlockStore
	txapp       TxApp
	accounts    Accounts
	validators  Validators
	snapshotter SnapshotModule

	// Broadcasters
	proposalBroadcaster     ProposalBroadcaster
	blkAnnouncer            BlkAnnouncer
	ackBroadcaster          AckBroadcaster
	blkRequester            BlkRequester
	rstStateBroadcaster     ResetStateBroadcaster
	discoveryReqBroadcaster DiscoveryReqBroadcaster
}

// ProposalBroadcaster broadcasts the new block proposal message to the network
type ProposalBroadcaster func(ctx context.Context, blk *types.Block)

// BlkAnnouncer broadcasts the new committed block to the network using the blockAnn message
type BlkAnnouncer func(ctx context.Context, blk *types.Block, appHash types.Hash)

// AckBroadcaster gossips the ack/nack messages to the network
type AckBroadcaster func(ack bool, height int64, blkID types.Hash, appHash *types.Hash) error

// BlkRequester requests the block from the network based on the height
type BlkRequester func(ctx context.Context, height int64) (types.Hash, types.Hash, []byte, error)

type ResetStateBroadcaster func(height int64) error

type DiscoveryReqBroadcaster func()

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

	consensusTx sql.PreparedTx

	blkProp  *blockProposal
	blockRes *blockResult
	lc       *lastCommit

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

type NetworkParams struct {
	// MaxBlockSize is the maximum size of a block in bytes.
	MaxBlockSize int64
	// JoinExpiry is the number of blocks after which the validators
	// join request expires if not approved.
	JoinExpiry int64
	// VoteExpiry is the default number of blocks after which the validators
	// vote expires if not approved.
	VoteExpiry int64
	// DisabledGasCosts dictates whether gas costs are disabled.
	DisabledGasCosts bool
	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64
}

type GenesisParams struct {
	ChainID string
	Params  *NetworkParams
}

// TODO: remove this
var defaultGenesisParams = &GenesisParams{
	ChainID: "test-chain",
	Params: &NetworkParams{
		DisabledGasCosts: true,
		JoinExpiry:       14400,
		VoteExpiry:       108000,
		MaxBlockSize:     6 * 1024 * 1024,
		MaxVotesPerTx:    200,
	},
}

// Config is the struct given to the constructor, [New].
type Config struct {
	// Signer is the private key of the node.
	PrivateKey crypto.PrivateKey
	// Leader is the public key of the leader.
	Leader crypto.PublicKey

	GenesisHash   types.Hash
	GenesisParams *GenesisParams // *config.GenesisConfig

	DB *pg.DB
	// Mempool is the mempool of the node.
	Mempool Mempool
	// BlockStore is the blockstore of the node.
	BlockStore BlockStore

	// ValidatorStore is the store of the validators.
	ValidatorStore Validators

	// Accounts is the store of the accounts.
	Accounts Accounts

	// SnapshotStore is the store of the snapshots.
	Snapshots SnapshotModule

	// TxApp is the transaction application layer.
	TxApp TxApp

	// ValidatorSet is the set of validators in the network.
	ValidatorSet map[string]ktypes.Validator
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

	// Determine *genesis* role based on leader pubkey and validator set.
	var role types.Role
	pubKey := cfg.PrivateKey.Public()

	if pubKey.Equals(cfg.Leader) {
		role = types.RoleLeader
		logger.Info("You are the leader")
	} else {
		pubKeyBts := pubKey.Bytes()
		if _, in := cfg.ValidatorSet[hex.EncodeToString(pubKeyBts)]; in {
			role = types.RoleValidator
			logger.Info("You are a validator")
		} else {
			role = types.RoleSentry
			logger.Info("You are a sentry")
		}
	}

	signer := auth.GetSigner(cfg.PrivateKey)

	if cfg.GenesisParams == nil {
		cfg.GenesisParams = defaultGenesisParams // TODO: remove
	}

	// rethink how this state is initialized
	ce := &ConsensusEngine{
		signer:         signer,
		pubKey:         pubKey,
		privKey:        cfg.PrivateKey,
		leader:         cfg.Leader,
		proposeTimeout: cfg.ProposeTimeout,
		db:             cfg.DB,
		state: state{
			blkProp:  nil,
			blockRes: nil,
			lc: &lastCommit{ // the zero values don't need to be specified, but for completeness...
				height:  0,
				blkHash: zeroHash,
				appHash: zeroHash,
			},
			votes: make(map[string]*vote),
		},
		stateInfo: StateInfo{
			height:  0,
			status:  Committed,
			blkProp: nil,
		},
		chainCtx: &common.ChainContext{
			ChainID: cfg.GenesisParams.ChainID,
			NetworkParameters: &common.NetworkParameters{
				MaxBlockSize:     cfg.GenesisParams.Params.MaxBlockSize,
				JoinExpiry:       cfg.GenesisParams.Params.JoinExpiry,
				VoteExpiry:       cfg.GenesisParams.Params.VoteExpiry,
				DisabledGasCosts: cfg.GenesisParams.Params.DisabledGasCosts,
				MaxVotesPerTx:    cfg.GenesisParams.Params.MaxVotesPerTx,
			},
			// MigrationParams:
		},
		validatorSet: maps.Clone(cfg.ValidatorSet),
		msgChan:      make(chan consensusMessage, 1), // buffer size??
		haltChan:     make(chan struct{}, 1),
		resetChan:    make(chan int64, 1),
		bestHeightCh: make(chan *discoveryMsg, 1),
		// interfaces
		mempool:        cfg.Mempool,
		blockStore:     cfg.BlockStore,
		txapp:          cfg.TxApp,
		accounts:       cfg.Accounts,
		validators:     cfg.ValidatorStore,
		snapshotter:    cfg.Snapshots,
		log:            logger,
		genesisAppHash: cfg.GenesisHash,
	}

	ce.role.Store(role)
	ce.networkHeight.Store(0)

	return ce
}

var initialHeight int64 = 0 // TODO: get it from genesis?

func (ce *ConsensusEngine) Start(ctx context.Context, proposerBroadcaster ProposalBroadcaster,
	blkAnnouncer BlkAnnouncer, ackBroadcaster AckBroadcaster, blkRequester BlkRequester, stateResetter ResetStateBroadcaster,
	discoveryReqBroadcaster DiscoveryReqBroadcaster) error {
	ce.proposalBroadcaster = proposerBroadcaster
	ce.blkAnnouncer = blkAnnouncer
	ce.ackBroadcaster = ackBroadcaster
	ce.blkRequester = blkRequester
	ce.rstStateBroadcaster = stateResetter
	ce.discoveryReqBroadcaster = discoveryReqBroadcaster

	ce.log.Info("Starting the consensus engine")

	// Fast catchup the node with the network height
	if err := ce.catchup(ctx); err != nil {
		return fmt.Errorf("error catching up: %w", err)
	}

	// start mining
	ce.startMining(ctx)

	// start the event loop
	return ce.runConsensusEventLoop(ctx)
}

// GenesisInit initializes the node with the genesis state. This included initializing the
// votestore with the genesis validators, accounts with the genesis allocations and the
// chain meta store with the genesis network parameters.
// This is called only once when the node is bootstrapping for the first time.
func (ce *ConsensusEngine) GenesisInit(ctx context.Context) error {
	genesisTx, err := ce.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer genesisTx.Rollback(ctx)

	// TODO: genesis allocs

	// genesis validators
	genVals := make([]*ktypes.Validator, 0, len(ce.validatorSet))
	for _, v := range ce.validatorSet {
		genVals = append(genVals, &ktypes.Validator{
			PubKey: v.PubKey,
			Power:  v.Power,
		})
	}

	startParams := *ce.chainCtx.NetworkParameters

	if err := ce.txapp.GenesisInit(ctx, genesisTx, genVals, nil, initialHeight, ce.chainCtx); err != nil {
		return err
	}

	if err := meta.SetChainState(ctx, genesisTx, initialHeight, ce.genesisAppHash[:], false); err != nil {
		return fmt.Errorf("error storing the genesis state: %w", err)
	}

	if err := meta.StoreDiff(ctx, genesisTx, &startParams, ce.chainCtx.NetworkParameters); err != nil {
		return fmt.Errorf("error storing the genesis consensus params: %w", err)
	}

	// TODO: Genesis hash and what are the mechanics for producing the first block (genesis block)?
	ce.txapp.Commit()

	ce.state.lc.appHash = ce.genesisAppHash
	ce.state.lc.height = initialHeight

	ce.log.Info("Initialized chain", "height", initialHeight, "appHash", ce.state.lc.appHash.String())
	return genesisTx.Commit(ctx)
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
func (ce *ConsensusEngine) runConsensusEventLoop(ctx context.Context) error {
	// TODO: make these configurable?
	catchUpTicker := time.NewTicker(5 * time.Second)
	reannounceTicker := time.NewTicker(3 * time.Second)
	blkPropTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			ce.log.Info("Shutting down the consensus engine")
			return nil

		case <-ce.haltChan:
			// Halt the network
			ce.resetState(ctx) // rollback the current block execution and stop the node
			ce.log.Error("Received halt signal, stopping the consensus engine")
			return nil

		case <-catchUpTicker.C:
			err := ce.doCatchup(ctx) // better name??
			if err != nil {
				panic(err) // TODO: should we panic here?
			}

		case <-reannounceTicker.C:
			ce.reannounceMsgs(ctx)

		case height := <-ce.resetChan:
			ce.resetBlockProp(height)

		case m := <-ce.msgChan:
			ce.handleConsensusMessages(ctx, m)

		case <-blkPropTicker.C:
			ce.rebroadcastBlkProposal(ctx)
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
	ce.log.Info("Consensus message received", "type", msg.MsgType, "sender", hex.EncodeToString(msg.Sender))

	switch v := msg.Msg.(type) {
	case *blockProposal:
		ce.processBlockProposal(ctx, v)

	case *vote:
		if ce.role.Load() != types.RoleLeader {
			return
		}
		if err := ce.addVote(ctx, v, hex.EncodeToString(msg.Sender)); err != nil {
			ce.log.Error("Error adding vote", "vote", v, "error", err)
			return
		}

	case *blockAnnounce:
		if err := ce.commitBlock(v.blk, v.appHash); err != nil {
			ce.log.Error("Error processing committing block", "error", err)
			return
		}

	default:
		ce.log.Warnf("Invalid message type received")
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

	ce.log.Info("Reset msg: ", "height", height)
	if ce.state.lc.height == height {
		if ce.state.blkProp != nil {
			ce.log.Info("Resetting the block proposal", "height", height)
			if err := ce.resetState(context.Background()); err != nil {
				ce.log.Error("Error resetting the state", "error", err) // panic? or consensus error?
			}
		}
	}
}

// catchup syncs the node first with the local blockstore and then with the network.
func (ce *ConsensusEngine) catchup(ctx context.Context) error {
	// Figure out the app state and initialize the node state.
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	// set the node to catchup mode
	ce.inSync.Store(true)

	readTx, err := ce.db.BeginReadTx(ctx)
	if err != nil {
		return err
	}
	defer readTx.Rollback(ctx)

	// retrieve the last committed block info from the blockstore
	storeHeight, blkHash, storeAppHash := ce.blockStore.Best()

	// retrieve the app state from the meta table
	appHeight, appHash, dirty, err := meta.GetChainState(ctx, readTx)
	if err != nil {
		return err
	}

	if dirty {
		ce.log.Info("App state is dirty, partially committed??") // TODO: what to be done here??
	}

	ce.log.Info("Initial Node state: ", "appHeight", appHeight, "storeHeight", storeHeight, "appHash", appHash, "storeAppHash", storeAppHash)

	if appHeight > storeHeight {
		// This is not possible, App can't be ahead of the store
		return fmt.Errorf("app height %d is greater than the store height %d", appHeight, storeHeight)
	}

	if appHeight == storeHeight && !bytes.Equal(appHash, storeAppHash[:]) {
		// This is not possible, PG mismatches with the Blockstore return error
		return fmt.Errorf("AppHash mismatch, appHash: %x, storeAppHash: %x", appHash, storeAppHash)
	}

	if appHeight > 0 {
		ce.setLastCommitInfo(appHeight, blkHash, types.Hash(appHash))
	}

	if appHeight == -1 {
		// This is the first time the node is bootstrapping
		// initialize the db with the genesis state
		if err := ce.GenesisInit(ctx); err != nil {
			return fmt.Errorf("error initializing the genesis state: %w", err)
		}
	}

	// Replay the blocks from the blockstore if the app hasn't played all the blocks yet.
	if appHeight < storeHeight {
		if err := ce.replayFromBlockStore(appHeight+1, storeHeight); err != nil {
			return err
		}
	}

	// Sync with the network using the blocksync
	if err := ce.doBlockSync(ctx); err != nil {
		return err
	}

	// Done with the catchup
	ce.inSync.Store(false)

	return nil
}

func (ce *ConsensusEngine) setLastCommitInfo(height int64, blkHash types.Hash, appHash types.Hash) {
	ce.state.lc.height = height
	ce.state.lc.appHash = appHash
	ce.state.lc.blkHash = blkHash
	// TODO: do we need to set the block ???
	// ce.state.lc.blk = nil

	ce.stateInfo.height = height
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
}

// replayBlocks replays all the blocks from the blockstore if the app hasn't played all the blocks yet.
func (ce *ConsensusEngine) replayFromBlockStore(startHeight, bestHeight int64) error {
	height := startHeight
	t0 := time.Now()

	if startHeight <= bestHeight {
		return nil // already caught up with the blockstore
	}

	for height <= bestHeight {
		_, blk, appHash, err := ce.blockStore.GetByHeight(height)
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

		height++
	}

	ce.log.Info("Replayed blocks from the blockstore", "from", startHeight, "to", height, "elapsed", time.Since(t0), "appHash", ce.state.lc.appHash)
	return nil
}

// Blocksync need to be way quicker, whereas the others need not be that frequent.
func (ce *ConsensusEngine) reannounceMsgs(ctx context.Context) {
	// Leader should reannounce the blkProp and blkAnn messages
	// Validators should reannounce the Ack messages
	ce.state.mtx.RLock()
	defer ce.state.mtx.RUnlock()

	if ce.role.Load() == types.RoleLeader && ce.state.lc.height > 0 {
		// Announce block commit message for the last committed block
		if ce.state.lc.blk != nil {
			go ce.blkAnnouncer(ctx, ce.state.lc.blk, ce.state.lc.appHash) // TODO: can be made infrequent
		}
		return
	}

	if ce.role.Load() == types.RoleValidator {
		// reannounce the acks, if still waiting for the commit message
		if ce.state.blkProp != nil && ce.state.blockRes != nil &&
			!ce.state.blockRes.appHash.IsZero() && ce.networkHeight.Load() <= ce.state.lc.height {
			ce.log.Info("Reannouncing ACK", "ack", ce.state.blockRes.ack, "height", ce.state.blkProp.height, "hash", ce.state.blkProp.blkHash)
			go ce.ackBroadcaster(ce.state.blockRes.ack, ce.state.blkProp.height, ce.state.blkProp.blkHash, &ce.state.blockRes.appHash)
		}
	}
}

func (ce *ConsensusEngine) rebroadcastBlkProposal(ctx context.Context) {
	ce.state.mtx.RLock()
	defer ce.state.mtx.RUnlock()

	if ce.role.Load() == types.RoleLeader && ce.state.blkProp != nil {
		go ce.proposalBroadcaster(ctx, ce.state.blkProp.blk)
	}
}

func (ce *ConsensusEngine) doCatchup(ctx context.Context) error {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.role.Load() == types.RoleLeader {
		return nil
	}

	if ce.state.lc.height >= ce.networkHeight.Load() {
		return nil
	}

	startHeight := ce.state.lc.height + 1
	t0 := time.Now()

	if ce.role.Load() == types.RoleValidator {
		// If validator is in the middle of processing a block, finish it first

		if ce.state.blkProp != nil && ce.state.blockRes != nil { // Waiting for the commit message
			blkHash, appHash, rawBlk, err := ce.blkRequester(ctx, ce.state.blkProp.height)
			if err != nil {
				ce.log.Warn("Error requesting block from network", "height", ce.state.blkProp.height, "error", err)
				return nil // not an error, just retry later
			}

			if blkHash != ce.state.blkProp.blkHash { // processed incorrect block
				if err := ce.resetState(ctx); err != nil {
					return fmt.Errorf("error aborting incorrect block execution: height: %d, blkID: %x, error: %w", ce.state.blkProp.height, blkHash, err)
				}

				blk, err := types.DecodeBlock(rawBlk)
				if err != nil {
					return fmt.Errorf("failed to decode the block, blkHeight: %d, blkID: %x, error: %w", ce.state.blkProp.height, blkHash, err)
				}

				if err := ce.processAndCommit(blk, appHash); err != nil {
					return fmt.Errorf("failed to replay the block: blkHeight: %d, blkID: %x, error: %w", ce.state.blkProp.height, blkHash, err)
				}
			} else {
				if appHash == ce.state.blockRes.appHash {
					// commit the block
					if err := ce.commit(); err != nil {
						return fmt.Errorf("failed to commit the block: height: %d, error: %w", ce.state.blkProp.height, err)
					}

					ce.nextState()
				} else {
					// halt the network
					ce.log.Error("Incorrect AppHash, halting the node.", "received", appHash, "has", ce.state.blockRes.appHash)
					close(ce.haltChan)
				}
			}
		}
	}

	err := ce.replayBlockFromNetwork(ctx)
	if err != nil {
		return err
	}

	ce.log.Info("Network Sync: ", "from", startHeight, "to (excluding)", ce.state.lc.height+1, "time", time.Since(t0), "appHash", ce.state.lc.appHash)

	return nil
}

func (ce *ConsensusEngine) Role() types.Role {
	return ce.role.Load().(types.Role)
}

func (ce *ConsensusEngine) updateNetworkHeight(height int64) {
	if height > ce.networkHeight.Load() {
		ce.networkHeight.Store(height)
	}
}

func (ce *ConsensusEngine) hasMajorityCeil(cnt int) bool {
	threshold := len(ce.validatorSet)/2 + 1 // majority votes required
	return cnt >= threshold
}

func (ce *ConsensusEngine) hasMajorityFloor(cnt int) bool {
	threshold := len(ce.validatorSet) / 2
	return cnt >= threshold
}
