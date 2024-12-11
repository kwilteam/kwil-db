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
	validatorSet   map[string]ktypes.Validator // key: hex encoded pubkey
	genesisAppHash types.Hash

	// stores state machine state for the consensus engine
	state  state
	inSync atomic.Bool // set when the node is still catching up with the network during bootstrapping

	// copy of the minimal state info for the p2p layer usage.
	stateInfo StateInfo

	// Channels
	newRound     chan struct{}
	msgChan      chan consensusMessage
	haltChan     chan struct{}      // can take a msg or reason for halting the network
	resetChan    chan int64         // to reset the state of the consensus engine
	bestHeightCh chan *discoveryMsg // to sync the leader with the network

	// interfaces
	db             DB
	mempool        Mempool
	blockStore     BlockStore
	blockProcessor BlockProcessor

	// protects the mempool access. Commit takes this lock to ensure that
	// no new txs are added to the mempool while the block is being committed
	// i.e while the accounts are being updated.
	mempoolMtx sync.Mutex

	// Broadcasters
	proposalBroadcaster     ProposalBroadcaster
	blkAnnouncer            BlkAnnouncer
	ackBroadcaster          AckBroadcaster
	blkRequester            BlkRequester
	rstStateBroadcaster     ResetStateBroadcaster
	discoveryReqBroadcaster DiscoveryReqBroadcaster

	// waitgroup to track all the consensus goroutines
	wg sync.WaitGroup
}

// Config is the struct given to the constructor, [New].
type Config struct {
	// Signer is the private key of the node.
	PrivateKey crypto.PrivateKey
	// Leader is the public key of the leader.
	Leader crypto.PublicKey

	GenesisHash types.Hash

	DB *pg.DB
	// Mempool is the mempool of the node.
	Mempool Mempool
	// BlockStore is the blockstore of the node.
	BlockStore BlockStore

	BlockProcessor BlockProcessor

	// ValidatorSet is the set of validators in the network.
	ValidatorSet map[string]ktypes.Validator
	// Logger is the logger of the node.
	Logger log.Logger

	// ProposeTimeout is the timeout for proposing a block.
	ProposeTimeout time.Duration
}

// ProposalBroadcaster broadcasts the new block proposal message to the network
type ProposalBroadcaster func(ctx context.Context, blk *ktypes.Block)

// BlkAnnouncer broadcasts the new committed block to the network using the blockAnn message
type BlkAnnouncer func(ctx context.Context, blk *ktypes.Block, appHash types.Hash)

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
	blk     *ktypes.Block // why is this needed? can be fetched from the blockstore too.
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
		validatorSet: maps.Clone(cfg.ValidatorSet),
		msgChan:      make(chan consensusMessage, 1), // buffer size??
		haltChan:     make(chan struct{}, 1),
		resetChan:    make(chan int64, 1),
		bestHeightCh: make(chan *discoveryMsg, 1),
		newRound:     make(chan struct{}, 1),
		// interfaces
		mempool:        cfg.Mempool,
		blockStore:     cfg.BlockStore,
		blockProcessor: cfg.BlockProcessor,
		log:            logger,
		genesisAppHash: cfg.GenesisHash,
	}

	ce.role.Store(role)
	ce.networkHeight.Store(0)

	return ce
}

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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fast catchup the node with the network height
	if err := ce.catchup(ctx); err != nil {
		return fmt.Errorf("error catching up: %w", err)
	}

	// start mining
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()

		ce.startMining(ctx)
	}()

	// start the event loop
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		defer cancel() // stop CE in case event loop terminated early e.g. halt

		ce.runConsensusEventLoop(ctx)
		ce.log.Info("Consensus event loop stopped...")
	}()

	ce.wg.Wait()

	ce.close()
	ce.log.Info("Consensus engine stopped")
	return nil
}

func (ce *ConsensusEngine) close() {
	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if err := ce.blockProcessor.Close(); err != nil {
		ce.log.Error("Error closing the block processor", "error", err)
	}
}

// runEventLoop starts the event loop for the consensus engine.
// Below are the external event triggers that nodes can receive depending on their role:
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
	ce.log.Info("Starting the consensus event loop...")
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
			// ce.resetState(ctx) // rollback the current block execution and stop the node
			ce.log.Error("Received halt signal, stopping the consensus engine")
			return nil

		case <-ce.newRound:
			if err := ce.startNewRound(ctx); err != nil {
				ce.log.Error("Error starting a new round", "error", err)
				return err
			}

		case <-catchUpTicker.C:
			err := ce.doCatchup(ctx) // better name??
			if err != nil {
				panic(err) // TODO: should we panic here?
			}

		case <-reannounceTicker.C:
			ce.reannounceMsgs(ctx)

		case height := <-ce.resetChan:
			ce.resetBlockProp(ctx, height)

		case m := <-ce.msgChan:
			ce.handleConsensusMessages(ctx, m)

		case <-blkPropTicker.C:
			ce.rebroadcastBlkProposal(ctx)
		}
	}
}

// startMining starts the mining process based on the role of the node.
func (ce *ConsensusEngine) startMining(_ context.Context) error {
	// Start the mining process if the node is a leader
	// validators and sentry nodes get activated when they receive a block proposal or block announce msgs.
	if ce.role.Load() == types.RoleLeader {
		ce.log.Infof("Starting the leader node")
		ce.newRound <- struct{}{}
	} else {
		ce.log.Infof("Starting the validator/sentry node")
	}

	return nil
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
		if err := ce.commitBlock(ctx, v.blk, v.appHash); err != nil {
			ce.log.Error("Error processing committing block", "error", err)
			return
		}

	default:
		ce.log.Warnf("Invalid message type received")
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
		return fmt.Errorf("app height %d is greater than the store height %d (did you forget to reset postgres?)", appHeight, storeHeight)
	}

	if appHeight == -1 {
		// This is the first time the node is bootstrapping
		// initialize the db with the genesis state
		genHeight, genAppHash, err := ce.blockProcessor.InitChain(ctx)
		if err != nil {
			return fmt.Errorf("error initializing the chain: %w", err)
		}

		ce.state.lc.height = genHeight
		copy(ce.state.lc.appHash[:], genAppHash)

	} else if appHeight > 0 {
		if appHeight == storeHeight && !bytes.Equal(appHash, storeAppHash[:]) {
			// This is not possible, PG mismatches with the Blockstore return error
			return fmt.Errorf("AppHash mismatch, appHash: %x, storeAppHash: %v", appHash, storeAppHash)
		}

		ce.setLastCommitInfo(appHeight, blkHash, types.Hash(appHash))
	}

	// Replay the blocks from the blockstore if the app hasn't played all the blocks yet.
	if appHeight < storeHeight {
		if err := ce.replayFromBlockStore(ctx, appHeight+1, storeHeight); err != nil {
			return err
		}
	}

	// Sync with the network using the blocksync
	if err := ce.doBlockSync(ctx); err != nil {
		return err
	}

	// Update the role of the node based on the final state of the validator set
	ce.updateValidatorSetAndRole()

	// Done with the catchup
	ce.inSync.Store(false)

	return nil
}

// updateRole updates the validator set and role of the node based on the final state of the validator set.
// This is called at the end of the commit phase or at the end of the catchup phase during bootstrapping.
func (ce *ConsensusEngine) updateValidatorSetAndRole() error {
	valset := ce.blockProcessor.GetValidators()
	pubKey := ce.privKey.Public()

	ce.validatorSet = make(map[string]ktypes.Validator)
	for _, v := range valset {
		ce.validatorSet[hex.EncodeToString(v.PubKey)] = ktypes.Validator{
			PubKey: v.PubKey,
			Power:  v.Power,
		}
	}

	currentRole := ce.role.Load()

	if pubKey.Equals(ce.leader) {
		ce.role.Store(types.RoleLeader)
		return nil
	}

	_, ok := ce.validatorSet[hex.EncodeToString(pubKey.Bytes())]
	if ok {
		ce.role.Store(types.RoleValidator)
	} else {
		ce.role.Store(types.RoleSentry)
	}

	finalRole := ce.role.Load()

	if currentRole != finalRole {
		ce.log.Info("Role updated", "from", currentRole, "to", finalRole)
	}
	return nil
}

func (ce *ConsensusEngine) setLastCommitInfo(height int64, blkHash types.Hash, appHash types.Hash) {
	ce.state.lc.height = height
	ce.state.lc.appHash = appHash
	ce.state.lc.blkHash = blkHash

	ce.stateInfo.height = height
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
}

// replayBlocks replays all the blocks from the blockstore if the app hasn't played all the blocks yet.
func (ce *ConsensusEngine) replayFromBlockStore(ctx context.Context, startHeight, bestHeight int64) error {
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

		err = ce.processAndCommit(ctx, blk, appHash)
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
			!ce.state.blockRes.appHash.IsZero() && ce.networkHeight.Load() <= ce.state.lc.height && ce.state.lc.height != 0 {
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
					return fmt.Errorf("error aborting incorrect block execution: height: %d, blkID: %v, error: %w", ce.state.blkProp.height, blkHash, err)
				}

				blk, err := ktypes.DecodeBlock(rawBlk)
				if err != nil {
					return fmt.Errorf("failed to decode the block, blkHeight: %d, blkID: %v, error: %w", ce.state.blkProp.height, blkHash, err)
				}

				if err := ce.processAndCommit(ctx, blk, appHash); err != nil {
					return fmt.Errorf("failed to replay the block: blkHeight: %d, blkID: %v, error: %w", ce.state.blkProp.height, blkHash, err)
				}
			} else {
				if appHash == ce.state.blockRes.appHash {
					// commit the block
					if err := ce.commit(ctx); err != nil {
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

	ce.updateValidatorSetAndRole()

	return nil
}

// resetBlockProp aborts the block execution and resets the state to the last committed block.
func (ce *ConsensusEngine) resetBlockProp(ctx context.Context, height int64) {
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
			if err := ce.resetState(ctx); err != nil {
				ce.log.Error("Error resetting the state", "error", err) // panic? or consensus error?
			}
		}
	}
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
