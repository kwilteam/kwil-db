package consensus

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	blockprocessor "github.com/kwilteam/kwil-db/node/block_processor"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/types"
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
	privKey crypto.PrivateKey
	pubKey  crypto.PublicKey
	log     log.Logger

	// proposeTimeout specifies the time duration to wait before proposing a new block for the next height.
	// This timeout is used by the leader to propose a block if transactions are available. Default is 1 second.
	proposeTimeout time.Duration
	// emptyBlockTimeout specifies the time duration to wait before proposing an empty block.
	// This should always be greater than the proposeTimeout. Default is 1 minute.
	emptyBlockTimeout time.Duration

	// blkProposalInterval specifies the time duration to wait before reannouncing the block proposal message.
	// This is only applicable for the leader. This timeout influences how quickly the out-of-sync nodes can
	// catchup with the consensus rounds, thereby influencing the block time. Default is 1 second.
	blkProposalInterval time.Duration

	// blkAnnReannounceTimeout specifies the time duration to wait before reannouncing the block announce message.
	blkAnnInterval time.Duration

	// broadcastTxTimeout specifies the time duration to wait for a transaction to be included in the block.
	broadcastTxTimeout time.Duration

	genesisHeight int64                       // height of the genesis block
	leader        crypto.PublicKey            // TODO: update with network param updates touching it
	validatorSet  map[string]ktypes.Validator // key: hex encoded pubkey

	// stores state machine state for the consensus engine
	state  state
	inSync atomic.Bool // set when the node is still catching up with the network during bootstrapping

	// copy of the minimal state info for the p2p layer usage.
	stateInfo StateInfo

	cancelFnMtx     sync.Mutex // protects blkExecCancelFn, longRunningTxs and numResets
	blkExecCancelFn context.CancelFunc
	// list of txs to be removed from the mempool
	// only used by the leader and protected by the cancelFnMtx
	longRunningTxs []ktypes.Hash
	numResets      int64

	// Channels
	newBlockProposal chan struct{} // triggers block production in the leader
	// newRound triggers the start of a new round in the consensus engine.
	// leader waits for the minBlockInterval before proposing a new block if txs are available.
	// if no txs are available for the maxBlockInterval duration, the leader proposes an empty block.
	newRound     chan struct{}
	msgChan      chan consensusMessage
	haltChan     chan string        // can take a msg or reason for halting the network
	resetChan    chan *resetMsg     // to reset the state of the consensus engine
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
	txAnnouncer             TxAnnouncer

	// TxSubscriber
	subMtx        sync.Mutex // protects access to txSubscribers
	txSubscribers map[ktypes.Hash]chan ktypes.TxResult

	// waitgroup to track all the consensus goroutines
	wg sync.WaitGroup
}

// Config is the struct given to the constructor, [New].
type Config struct {
	// Signer is the private key of the node.
	PrivateKey crypto.PrivateKey
	// Leader is the public key of the leader.
	Leader crypto.PublicKey
	// GenesisHeight is the initial height of the network.
	GenesisHeight int64

	// ProposeTimeout is the minimum time duration to wait before proposing a new block.
	// Leader can propose a block with transactions as soon as this timeout is reached. Default is 1 second.
	ProposeTimeout time.Duration
	// EmptyBlockTimeout is the maximum time duration to wait before proposing a new block without transactions.
	// Default is 1 minute.
	EmptyBlockTimeout time.Duration
	// BlkPropReannounceInterval is the frequency at which block proposal messages are reannounced by the Leader.
	BlockProposalInterval time.Duration
	// BlkAnnReannounceInterval is the frequency at which block commit messages are reannounced by the Leader.
	// This is also the frequency at which the validators reannounce the ack messages.
	BlockAnnInterval time.Duration
	// CatchUpInterval is the frequency at which the node attempts to catches up with the network if lagging.
	// CatchUpInterval  time.Duration
	BroadcastTxTimeout time.Duration

	// Interfaces
	DB             *pg.DB
	Mempool        Mempool
	BlockStore     BlockStore
	BlockProcessor BlockProcessor
	Logger         log.Logger
}

type BroadcastFns struct {
	ProposalBroadcaster     ProposalBroadcaster
	TxAnnouncer             TxAnnouncer
	BlkAnnouncer            BlkAnnouncer
	AckBroadcaster          AckBroadcaster
	BlkRequester            BlkRequester
	RstStateBroadcaster     ResetStateBroadcaster
	DiscoveryReqBroadcaster DiscoveryReqBroadcaster
	TxBroadcaster           blockprocessor.BroadcastTxFn
}

type WhitelistFns struct {
	AddPeer    func(string) error
	RemovePeer func(string) error

	// List func() []string
}

// ProposalBroadcaster broadcasts the new block proposal message to the network
type ProposalBroadcaster func(ctx context.Context, blk *ktypes.Block)

// BlkAnnouncer broadcasts the new committed block to the network using the blockAnn message
type BlkAnnouncer func(ctx context.Context, blk *ktypes.Block, ci *ktypes.CommitInfo)

// TxAnnouncer broadcasts the new transaction to the network
type TxAnnouncer func(ctx context.Context, txHash types.Hash, rawTx []byte)

// AckBroadcaster gossips the ack/nack messages to the network
type AckBroadcaster func(ack bool, height int64, blkID types.Hash, appHash *types.Hash, Signature []byte) error

// BlkRequester requests the block from the network based on the height
type BlkRequester func(ctx context.Context, height int64) (types.Hash, []byte, *ktypes.CommitInfo, error)

type ResetStateBroadcaster func(height int64, txIDs []ktypes.Hash) error

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

	lastCommit lastCommit
}

// Consensus state that is applicable for processing the block at a specific height.
type state struct {
	mtx sync.RWMutex

	blkProp  *blockProposal
	blockRes *blockResult
	lc       *lastCommit

	// Votes: Applicable only to the leader
	// These are the Acks received from the validators.
	votes map[string]*ktypes.VoteInfo

	commitInfo *ktypes.CommitInfo
}

type blockResult struct {
	ack          bool
	appHash      ktypes.Hash
	txResults    []ktypes.TxResult
	vote         *vote
	paramUpdates ktypes.ParamUpdates
}

type lastCommit struct {
	height  int64
	blkHash types.Hash

	appHash types.Hash

	blk        *ktypes.Block // for reannounce and other status getters
	commitInfo *ktypes.CommitInfo
}

// New creates a new consensus engine.
func New(cfg *Config) *ConsensusEngine {
	logger := cfg.Logger
	if logger == nil {
		// logger = log.DiscardLogger // for prod
		logger = log.New(log.WithName("CONS"), log.WithLevel(log.LevelDebug),
			log.WithWriter(os.Stdout), log.WithFormat(log.FormatUnstructured))
	}

	// defer role assignment till the beginning of the catchup phase.
	pubKey := cfg.PrivateKey.Public()

	// rethink how this state is initialized
	ce := &ConsensusEngine{
		pubKey:              pubKey,
		privKey:             cfg.PrivateKey,
		leader:              cfg.Leader,
		proposeTimeout:      cfg.ProposeTimeout,
		emptyBlockTimeout:   cfg.EmptyBlockTimeout,
		blkProposalInterval: cfg.BlockProposalInterval,
		blkAnnInterval:      cfg.BlockAnnInterval,
		broadcastTxTimeout:  cfg.BroadcastTxTimeout,
		db:                  cfg.DB,
		state: state{
			blkProp:  nil,
			blockRes: nil,
			lc: &lastCommit{ // the zero values don't need to be specified, but for completeness...
				height:  0,
				blkHash: zeroHash,
				appHash: zeroHash,
			},
			votes: make(map[string]*ktypes.VoteInfo),
		},
		stateInfo: StateInfo{
			height:  0,
			status:  Committed,
			blkProp: nil,
		},
		genesisHeight:    cfg.GenesisHeight,
		msgChan:          make(chan consensusMessage, 1), // buffer size??
		haltChan:         make(chan string, 1),
		resetChan:        make(chan *resetMsg, 1),
		bestHeightCh:     make(chan *discoveryMsg, 1),
		newRound:         make(chan struct{}, 1),
		newBlockProposal: make(chan struct{}, 1),

		// interfaces
		mempool:        cfg.Mempool,
		blockStore:     cfg.BlockStore,
		blockProcessor: cfg.BlockProcessor,
		log:            logger,
		txSubscribers:  make(map[ktypes.Hash]chan ktypes.TxResult),
	}

	// set it to sentry by default, will be updated in the catchup phase when the engine starts.
	// Not initializing the role here will panic the node, as a lot of RPC calls such as HealthCheck,
	// Status, etc. tries to access the role.
	ce.role.Store(types.RoleSentry)

	return ce
}

func (ce *ConsensusEngine) Start(ctx context.Context, fns BroadcastFns, peerFns WhitelistFns) error {
	ce.proposalBroadcaster = fns.ProposalBroadcaster
	ce.blkAnnouncer = fns.BlkAnnouncer
	ce.ackBroadcaster = fns.AckBroadcaster
	ce.blkRequester = fns.BlkRequester
	ce.rstStateBroadcaster = fns.RstStateBroadcaster
	ce.discoveryReqBroadcaster = fns.DiscoveryReqBroadcaster
	ce.txAnnouncer = fns.TxAnnouncer

	ce.blockProcessor.SetCallbackFns(fns.TxBroadcaster, peerFns.AddPeer, peerFns.RemovePeer)

	ce.log.Info("Starting the consensus engine")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fast catchup the node with the network height
	if err := ce.catchup(ctx); err != nil {
		return fmt.Errorf("error catching up: %w", err)
	}

	// Start the mining process if the node is a leader. Validators and sentry
	// nodes are activated when they receive a block proposal or block announce msg.
	if ce.role.Load() == types.RoleLeader {
		ce.log.Infof("Starting the leader node")
		ce.newBlockProposal <- struct{}{} // recv by runConsensusEventLoop, buffered
	} else {
		ce.log.Infof("Starting the validator/sentry node")
	}

	// start the event loop
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()
		defer cancel() // stop CE in case event loop terminated early e.g. halt

		ce.runConsensusEventLoop(ctx) // error return not needed?
		ce.log.Info("Consensus event loop stopped...")
	}()

	// resetChan listener
	ce.wg.Add(1)
	go func() {
		defer ce.wg.Done()

		ce.resetEventLoop(ctx)
		ce.log.Info("Reset event loop stopped...")
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

func (ce *ConsensusEngine) Status() *ktypes.NodeStatus {
	params := ce.blockProcessor.ConsensusParams()
	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()
	lc := ce.stateInfo.lastCommit
	var hdr *ktypes.BlockHeader
	if lc.blk != nil {
		hdr = lc.blk.Header
	}
	return &ktypes.NodeStatus{
		Role:            ce.role.Load().(types.Role).String(),
		CatchingUp:      ce.inSync.Load(),
		CommittedHeader: hdr,
		CommitInfo:      lc.commitInfo,
		Params:          params,
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
	ce.log.Info("Starting the consensus event loop...")
	catchUpTicker := time.NewTicker(5 * time.Second)        // Should this be configurable??
	reannounceTicker := time.NewTicker(ce.blkAnnInterval)   // 3 secs (default)
	blkPropTicker := time.NewTicker(ce.blkProposalInterval) // 1 sec (default)

	for {
		select {
		case <-ctx.Done():
			ce.log.Info("Shutting down the consensus engine")
			return nil

		case halt := <-ce.haltChan:
			ce.log.Error("Received halt signal, stopping the consensus engine", "reason", halt)
			return nil

		case <-ce.newRound:
			go ce.newBlockRound(ctx)
		case <-ce.newBlockProposal:
			params := ce.blockProcessor.ConsensusParams()
			if params.MigrationStatus == ktypes.MigrationCompleted {
				ce.log.Info("Network halted due to migration, no more blocks will be produced")
			}

			if err := ce.proposeBlock(ctx); err != nil {
				ce.log.Error("Error starting a new round", "error", err)
				return err
			}
		case <-catchUpTicker.C:
			err := ce.doCatchup(ctx)
			if err != nil {
				return err
			}

		case <-reannounceTicker.C:
			ce.reannounceMsgs(ctx)

		case m := <-ce.msgChan:
			ce.handleConsensusMessages(ctx, m)

		case <-blkPropTicker.C:
			ce.rebroadcastBlkProposal(ctx)
		}
	}
}

// resetEventLoop listens for the reset event and rollbacks the current block processing.
func (ce *ConsensusEngine) resetEventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			ce.log.Info("Shutting down the reset event loop")
			return
		case msg := <-ce.resetChan:
			ce.resetBlockProp(ctx, msg.height, msg.txIDs)
		}
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
			ce.log.Warn("Error adding vote", "vote", v, "error", err)
			return
		}

	case *blockAnnounce:
		if err := ce.commitBlock(ctx, v.blk, v.ci); err != nil {
			ce.log.Error("Error processing committed block", "error", err)
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
	storeHeight, blkHash, storeAppHash, _ /* timestamp */ := ce.blockStore.Best()

	// retrieve the app state from the meta table
	appHeight, appHash, dirty, err := meta.GetChainState(ctx, readTx)
	if err != nil {
		return err
	}

	if dirty {
		return fmt.Errorf("app state is dirty, error in the blockprocessor initialization, height: %d, appHash: %x", appHeight, appHash)
	}

	ce.log.Info("Initial Node state: ", "appHeight", appHeight, "storeHeight", storeHeight, "appHash", appHash, "storeAppHash", storeAppHash)

	if appHeight > storeHeight && appHeight != ce.genesisHeight {
		// This is not possible, App can't be ahead of the store
		return fmt.Errorf("app height %d is greater than the store height %d (did you forget to reset postgres?)", appHeight, storeHeight)
	}

	if appHeight == -1 {
		// This is the first time the node is bootstrapping
		// initialize the db with the genesis state
		appHeight, appHash, err = ce.blockProcessor.InitChain(ctx)
		if err != nil {
			return fmt.Errorf("error initializing the chain: %w", err)
		}

		ce.setLastCommitInfo(appHeight, nil, appHash)

	} else if appHeight > 0 {
		// restart or statesync init or zdt init
		if appHeight == storeHeight && !bytes.Equal(appHash, storeAppHash[:]) {
			// This is not possible, PG mismatches with the Blockstore return error
			return fmt.Errorf("AppHash mismatch, appHash: %x, storeAppHash: %v", appHash, storeAppHash)
		}
		ce.setLastCommitInfo(appHeight, blkHash[:], appHash)
	}

	// Set the role and validator set based on the initial state of the voters before starting the replay
	ce.updateValidatorSetAndRole()

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

	// Done with the catchup
	ce.inSync.Store(false)

	return nil
}

// updateRole updates the validator set and role of the node based on the final state of the validator set.
// This is called at the end of the commit phase or at the end of the catchup phase during bootstrapping.
func (ce *ConsensusEngine) updateValidatorSetAndRole() {
	valset := ce.blockProcessor.GetValidators()
	pubKey := ce.privKey.Public()

	ce.validatorSet = make(map[string]ktypes.Validator)
	for _, v := range valset {
		ce.validatorSet[hex.EncodeToString(v.Identifier)] = ktypes.Validator{
			AccountID: ktypes.AccountID{
				Identifier: v.Identifier,
				KeyType:    v.KeyType,
			},
			Power: v.Power,
		}
	}

	currentRole := ce.role.Load()

	if pubKey.Equals(ce.leader) {
		ce.role.Store(types.RoleLeader)
		return
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
}

func (ce *ConsensusEngine) setLastCommitInfo(height int64, blkHash []byte, appHash []byte) {
	ce.state.lc.height = height
	copy(ce.state.lc.appHash[:], appHash)
	copy(ce.state.lc.blkHash[:], blkHash)
	//
	// ce.state.lc.blk ?
	// ce.state.lc.commitInfo set in acceptCommitInfo (from commitBlock or processAndCommit)

	ce.stateInfo.height = height
	ce.stateInfo.status = Committed
	ce.stateInfo.blkProp = nil
}

// replayBlocks replays all the blocks from the blockstore if the app hasn't played all the blocks yet.
func (ce *ConsensusEngine) replayFromBlockStore(ctx context.Context, startHeight, bestHeight int64) error {
	height := startHeight
	t0 := time.Now()

	if startHeight >= bestHeight {
		return nil // already caught up with the blockstore
	}

	for height <= bestHeight {
		_, blk, ci, err := ce.blockStore.GetByHeight(height)
		if err != nil {
			if !errors.Is(err, types.ErrNotFound) {
				return fmt.Errorf("unexpected blockstore error: %w", err)
			}
			return nil // no more blocks to replay
		}

		err = ce.processAndCommit(ctx, blk, ci)
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
		if ce.state.lc.blk != nil && ce.state.commitInfo != nil {
			go ce.blkAnnouncer(ctx, ce.state.lc.blk, ce.state.lc.commitInfo)
		}
		return
	}

	if ce.role.Load() == types.RoleValidator {
		// reannounce the acks, if still waiting for the commit message
		if ce.state.blkProp != nil && ce.state.blockRes != nil &&
			!ce.state.blockRes.appHash.IsZero() {
			ce.log.Info("Reannouncing ACK", "ack", ce.state.blockRes.ack, "height", ce.state.blkProp.height, "hash", ce.state.blkProp.blkHash)
			vote := ce.state.blockRes.vote
			go ce.ackBroadcaster(vote.ack, vote.height, vote.blkHash, vote.appHash, vote.signature.Data)
		}
	}
}

func (ce *ConsensusEngine) rebroadcastBlkProposal(ctx context.Context) {
	ce.state.mtx.RLock()
	defer ce.state.mtx.RUnlock()

	if ce.role.Load() == types.RoleLeader && ce.state.blkProp != nil {
		ce.log.Info("Rebroadcasting block proposal", "height", ce.state.blkProp.height)
		go ce.proposalBroadcaster(ctx, ce.state.blkProp.blk)
	}
}

func (ce *ConsensusEngine) doCatchup(ctx context.Context) error {
	// status check, nodes halt here if the migration is completed
	params := ce.blockProcessor.ConsensusParams()
	if params.MigrationStatus == ktypes.MigrationCompleted {
		ce.log.Info("Network halted due to migration, no more blocks will be produced")
		return nil
	}

	ce.state.mtx.Lock()
	defer ce.state.mtx.Unlock()

	if ce.role.Load() == types.RoleLeader {
		return nil
	}

	startHeight := ce.state.lc.height + 1
	t0 := time.Now()

	if ce.role.Load() == types.RoleValidator {
		// If validator is in the middle of processing a block, finish it first

		if ce.state.blkProp != nil && ce.state.blockRes != nil { // Waiting for the commit message
			blkHash, rawBlk, ci, err := ce.blkRequester(ctx, ce.state.blkProp.height)
			if err != nil {
				ce.log.Warn("Error requesting block from network", "height", ce.state.blkProp.height, "error", err)
				return nil // not an error, just retry later
			}

			if blkHash != ce.state.blkProp.blkHash { // processed incorrect block
				if err := ce.rollbackState(ctx); err != nil {
					return fmt.Errorf("error aborting incorrect block execution: height: %d, blkID: %v, error: %w", ce.state.blkProp.height, blkHash, err)
				}

				blk, err := ktypes.DecodeBlock(rawBlk)
				if err != nil {
					return fmt.Errorf("failed to decode the block, blkHeight: %d, blkID: %v, error: %w", ce.state.blkProp.height, blkHash, err)
				}

				if err := ce.processAndCommit(ctx, blk, ci); err != nil {
					return fmt.Errorf("failed to replay the block: blkHeight: %d, blkID: %v, error: %w", ce.state.blkProp.height, blkHash, err)
				}
				// continue to replay blocks from network
			} else if ci.AppHash == ce.state.blockRes.appHash {
				// commit the block
				if err := ce.acceptCommitInfo(ci, blkHash); err != nil {
					return fmt.Errorf("failed to validate the commit info: height: %d, error: %w", ce.state.blkProp.height, err)
				}

				if err := ce.commit(ctx); err != nil {
					return fmt.Errorf("failed to commit the block: height: %d, error: %w", ce.state.blkProp.height, err)
				}

				ce.nextState()
			} else {
				// halt the network
				haltR := fmt.Sprintf("Incorrect AppHash, received: %v, have: %v", ci.AppHash, ce.state.blockRes.appHash)
				ce.haltChan <- haltR
			}
		}
	}

	err := ce.replayBlockFromNetwork(ctx)
	if err != nil {
		return err
	}

	ce.log.Info("Network Sync: ", "from", startHeight, "to", ce.state.lc.height, "time", time.Since(t0), "appHash", ce.state.lc.appHash)

	return nil
}

func (ce *ConsensusEngine) cancelBlock(height int64) bool {
	ce.log.Info("Reset msg: ", "height", height)

	ce.stateInfo.mtx.RLock()
	defer ce.stateInfo.mtx.RUnlock()

	// We will only honor the reset request if it's from the leader (already verified by now)
	// and the height is same as the last committed block height and the
	// block is still executing or waiting for the block commit message.
	if ce.stateInfo.height != height {
		ce.log.Warn("Invalid reset request", "height", height, "lastCommittedHeight", ce.state.lc.height)
		return false
	}

	if ce.stateInfo.blkProp == nil {
		ce.log.Info("Block already committed or executed, nothing to reset", "height", height)
		return false
	}

	ce.log.Info("Resetting the block proposal", "height", height)
	// cancel the context
	ce.cancelFnMtx.Lock()
	defer ce.cancelFnMtx.Unlock()
	if ce.blkExecCancelFn != nil {
		ce.blkExecCancelFn()
	}

	return true
}

// resetBlockProp aborts the block execution and resets the state to the last committed block.
func (ce *ConsensusEngine) resetBlockProp(ctx context.Context, height int64, txIDs []ktypes.Hash) {
	ce.log.Info("Reset msg: ", "height", height)

	rollback := ce.cancelBlock(height) // return here if false?

	// context is already cancelled, so try the lock
	ce.state.mtx.Lock() // block execution is cancelled by now
	defer ce.state.mtx.Unlock()

	// remove the long running txs from the mempool
	for _, txID := range txIDs {
		ce.mempool.Remove(txID)
	}

	if rollback {
		// rollback the state
		if err := ce.rollbackState(ctx); err != nil {
			ce.log.Error("Error aborting execution of block", "height", ce.state.blkProp.height, "blkID", ce.state.blkProp.blkHash, "error", err)
			return
		}
	}

	// recheck txs in the mempool, if we have deleted any txs from the mempool
	if len(txIDs) > 0 {
		ce.mempoolMtx.Lock()
		ce.mempool.RecheckTxs(ctx, ce.recheckTx)
		ce.mempoolMtx.Unlock()
	}
}

func (ce *ConsensusEngine) Role() types.Role {
	return ce.role.Load().(types.Role)
}

func (ce *ConsensusEngine) hasMajorityCeil(cnt int) bool {
	threshold := len(ce.validatorSet)/2 + 1 // majority votes required
	return cnt >= threshold
}

func (ce *ConsensusEngine) hasMajorityFloor(cnt int) bool {
	threshold := len(ce.validatorSet) / 2
	return cnt >= threshold
}

func (ce *ConsensusEngine) InCatchup() bool {
	return ce.inSync.Load()
}

func (ce *ConsensusEngine) SubscribeTx(txHash ktypes.Hash) (<-chan ktypes.TxResult, error) {
	ce.subMtx.Lock()
	defer ce.subMtx.Unlock()

	ch := make(chan ktypes.TxResult, 1)

	_, ok := ce.txSubscribers[txHash]
	if ok {
		return nil, fmt.Errorf("tx already subscribed")
	}

	ce.txSubscribers[txHash] = ch
	return ch, nil
}

func (ce *ConsensusEngine) UnsubscribeTx(txHash ktypes.Hash) {
	ce.subMtx.Lock()
	defer ce.subMtx.Unlock()

	delete(ce.txSubscribers, txHash)
}
