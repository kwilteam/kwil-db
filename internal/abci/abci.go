package abci

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/common/chain/forks"
	"github.com/kwilteam/kwil-db/common/ident"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	authExt "github.com/kwilteam/kwil-db/extensions/auth"
	"github.com/kwilteam/kwil-db/extensions/consensus"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
	"github.com/kwilteam/kwil-db/internal/abci/meta"
	"github.com/kwilteam/kwil-db/internal/statesync"
	"github.com/kwilteam/kwil-db/internal/txapp"
	"github.com/kwilteam/kwil-db/internal/version"
	parseCommon "github.com/kwilteam/kwil-db/parse/common"

	"go.uber.org/zap"
)

var (
	ABCIPeerFilterPath       = "/p2p/filter/"
	ABCIPeerFilterPathLen    = len(ABCIPeerFilterPath)
	statesyncSnapshotSchemas = []string{"kwild_voting", "kwild_internal", "kwild_chain", "kwild_accts", "kwild_migrations", "ds_*"}
	statsyncExcludedTables   = []string{"kwild_internal.sentry"}
	lastCommitInfoFile       = "last_commit_info.json"
)

// AbciConfig includes data that defines the chain and allow the application to
// satisfy the ABCI Application interface.
type AbciConfig struct {
	// GenesisAppHash is considered only when doing InitChain (genesis).
	GenesisAppHash     []byte
	ChainID            string
	ApplicationVersion uint64
	GenesisAllocs      map[string]*big.Int
	GasEnabled         bool
	ForkHeights        map[string]*uint64
	InitialHeight      int64

	ABCIDir string
}

type AbciApp struct {
	// db is a connection to the database
	db DB
	// consensusTx is the outermost transaction that wraps all other transactions
	// that can modify state. It should be set in FinalizeBlock and committed in Commit.
	consensusTx sql.PreparedTx
	// genesisTx is the transaction that is used at genesis, and in the first block.
	genesisTx sql.PreparedTx

	stateMtx sync.Mutex
	// appHash is the hash of the application state
	appHash []byte
	// height is the current block height
	height int64

	cfg   AbciConfig
	forks forks.Forks

	// snapshotter is the snapshotter module that handles snapshotting
	snapshotter SnapshotModule

	// replayingBlocks is a function that tells us whether we are in replay mode (syncing with the network),
	// or whether we are in normal operation mode.
	replayingBlocks func() bool

	// bootstrapper is the bootstrapper module that handles bootstrapping the database
	statesyncer StateSyncModule

	log log.Logger

	txApp TxApp

	consensusParams *chain.ConsensusParams

	chainContext    *common.ChainContext
	chainContextMtx sync.RWMutex

	broadcastFn EventBroadcaster

	// validatorAddressToPubKey is a map of validator addresses to their public
	// keys. It should only be accessed from consensus connection methods, which
	// are not called concurrently, or the constructor.
	validatorAddressToPubKey map[string][]byte

	// verifiedTxns stores hashes of all the transactions currently in the
	// mempool, which have passed signature verification. This is used to avoid
	// recomputing the hash for all mempool transactions on every TxQuery
	// request (to mitigate Potential DDOS attack vector).
	// https://github.com/kwilteam/kwil-db/issues/714
	verifiedTxnsMtx sync.RWMutex
	verifiedTxns    map[chainHash]struct{}

	// halted is set to true when the network is halted for migration.
	halted atomic.Bool

	// p2p is the whitelist of peers
	p2p WhitelistPeersModule

	// Migrator is the migrator module that handles migrations
	migrator MigratorModule

	// lastCommitInfoFileName is the file name of the last commit info file
	// which stores the app hash and height at the end of FinalizeBlock.
	lastCommitInfoFileName string
}

func NewAbciApp(ctx context.Context, cfg *AbciConfig, snapshotter SnapshotModule, statesyncer StateSyncModule,
	txRouter TxApp, consensusParams *chain.ConsensusParams, peers WhitelistPeersModule, migrator MigratorModule, db DB, logger log.Logger) (*AbciApp, error) {
	app := &AbciApp{
		db:                     db,
		cfg:                    *cfg,
		statesyncer:            statesyncer,
		snapshotter:            snapshotter,
		txApp:                  txRouter,
		migrator:               migrator,
		consensusParams:        consensusParams,
		appHash:                cfg.GenesisAppHash,
		p2p:                    peers,
		log:                    logger,
		lastCommitInfoFileName: filepath.Join(cfg.ABCIDir, lastCommitInfoFile),

		validatorAddressToPubKey: make(map[string][]byte),
		verifiedTxns:             make(map[chainHash]struct{}),
	}
	app.forks.FromMap(cfg.ForkHeights)

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin outer tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Populate the validatorAddressToPubKey field.
	validators, err := txRouter.GetValidators(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators: %w", err)
	}
	for _, val := range validators {
		addr, err := cometbft.PubkeyToAddr(val.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
		}

		app.validatorAddressToPubKey[addr] = val.PubKey
	}

	height, appHash, err := meta.GetChainState(ctx, tx)
	if err != nil {
		return nil, err
	}
	if height == -1 {
		height = 0 // negative means first start (no state table yet), but need non-negative for below logic
	}

	if bytes.Equal(appHash, []byte{0x42}) {
		// This is an interesting corner case where the node crashes after the commit of block execution
		// state but before the commit of the app hash. The chain state will have the app hash as "42".
		// CometBFT seems to persist the FinalizeBlock responses, so during restart it calls "Commit"
		// directly without calling FinalizeBlock and assumes that the application has the correct state
		// persisted, whereas the application has the app hash as "42". This will cause the current block
		// to be committed with an invalid app hash (any node doing a blocksync will fail at this point).
		// To fix this, we will store the last commit info in a file and use that as the application's
		// app hash whenever we find the chain state to be inconsistent signified by the app hash as "42".
		lc, err := loadLastCommitInfo(app.lastCommitInfoFileName)
		if err != nil {
			return nil, fmt.Errorf("failed to load last commit info due to error: %v. Please drop PostgresDB and rebuild state from existing block data", err)
		}

		if height != lc.Height {
			return nil, fmt.Errorf("height mismatch between chain state and last commit info: %d != %d. Please drop PostgresDB and rebuild state from existing block data", height, lc.Height)
		}

		appHash = lc.AppHash
		logger.Warn("Recovered last commit info", log.Int("height", height), log.String("appHash", hex.EncodeToString(appHash)))
	}

	app.appHash = appHash
	app.height = height

	app.log.Infof("Preparing ABCI application at height %v, appHash %x", height, appHash)

	// Enable any dynamically registered payloads, encoders, etc. from
	// extension-defined forks that must be enabled by the current height. In
	// addition to node restart, this is where forks with genesis height (0)
	// activation are enabled since the first FinalizeBlock is for height 1.
	activeForks := app.forks.ActivatedBy(uint64(height))
	slices.SortStableFunc(activeForks, forks.ForkSortFunc)
	for _, fork := range activeForks {
		forkName := fork.Name
		app.log.Infof("Hardfork %v activated at height %d", forkName, fork.Height)
		fork, have := consensus.Hardforks[forkName]
		if !have {
			return nil, fmt.Errorf("undefined hard fork %v", forkName)
		}

		// Update transaction payloads.
		for _, newPayload := range fork.TxPayloads {
			logger.Infof("Registering transaction route for payload type %v", fork.Name)
			if err := txapp.RegisterRouteImpl(newPayload.Type, newPayload.Route); err != nil {
				return nil, fmt.Errorf("failed to register route for payload %v: %w", newPayload.Type, err)
			}
		}
		// Update authenticators.
		for _, authMod := range fork.AuthUpdates {
			authExt.RegisterAuthenticator(authMod.Operation, authMod.Name, authMod.Authn)
		}
		// Update resolutions.
		for _, resMod := range fork.ResolutionUpdates {
			resolutions.RegisterResolution(resMod.Name, resMod.Operation, *resMod.Config)
		}
		// Update serialization codecs.
		for _, enc := range fork.Encoders {
			serialize.RegisterCodec(enc)
		}

		// NOTE: Forks defined with activation at genesis do NOT have their
		// consensus parameter updates or state mods applied. When specified
		// with activation height 0, the full desired consensus parameters
		// should be specified in genesis.json. When restarting above genesis
		// height, these updates would already have been applied by cometbft,
		// except for the kwil-specific parameters, which are loaded from the
		// ABCI and applied below.
	}

	var migrationParams *common.MigrationContext
	startHeight := app.consensusParams.Migration.StartHeight
	endHeight := app.consensusParams.Migration.EndHeight

	if startHeight != 0 && endHeight != 0 {
		migrationParams = &common.MigrationContext{
			StartHeight: startHeight,
			EndHeight:   endHeight,
		}
	}

	// if the network params have never been stored (which is the case for a fresh network),
	// we need to persist them for the first time. If they are found, we need to update our consensus params
	// with whatever the value is, since the consensus params here are read from the genesis file, and
	// may have been altered if this is not a new network.

	networkParams, err := meta.LoadParams(ctx, tx)
	if errors.Is(err, meta.ErrParamsNotFound) {
		status := types.NoActiveMigration
		if startHeight != 0 && endHeight != 0 {
			status = types.GenesisMigration
		}

		networkParams = &common.NetworkParameters{
			MaxBlockSize:     app.consensusParams.Block.MaxBytes,
			JoinExpiry:       app.consensusParams.Validator.JoinExpiry,
			VoteExpiry:       app.consensusParams.Votes.VoteExpiry,
			DisabledGasCosts: app.consensusParams.WithoutGasCosts,
			MaxVotesPerTx:    app.consensusParams.Votes.MaxVotesPerTx,
			MigrationStatus:  status,
		}

		// we need to store the genesis network params
		err = meta.StoreParams(ctx, tx, networkParams)
		if err != nil {
			return nil, fmt.Errorf("failed to store network params: %w", err)
		}

	} else if err != nil {
		return nil, fmt.Errorf("failed to load network params: %w", err)
	} else {
		// we will apply the netParams to the consensus params
		app.consensusParams.Block.MaxBytes = networkParams.MaxBlockSize
		app.consensusParams.Validator.JoinExpiry = networkParams.JoinExpiry
		app.consensusParams.Votes.VoteExpiry = networkParams.VoteExpiry
		app.consensusParams.WithoutGasCosts = networkParams.DisabledGasCosts
		app.consensusParams.Votes.MaxVotesPerTx = networkParams.MaxVotesPerTx
	}

	app.chainContext = &common.ChainContext{
		ChainID:           cfg.ChainID,
		NetworkParameters: networkParams,
		MigrationParams:   migrationParams,
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit outer tx: %w", err)
	}

	return app, nil
}

// proposerAddrToString converts a proposer address to a string.
// This follows the semantics of comet's ed25519.Pubkey.Address() method,
// which hex encodes and upper cases the address
func proposerAddrToString(addr []byte) string {
	return strings.ToUpper(hex.EncodeToString(addr))
}

type chainHash = [32]byte

func (a *AbciApp) ChainID() string {
	return a.cfg.ChainID
}

// Activations consults chain config for the names of hard forks that activate
// at the given block height, and retrieves the associated changes from the
// consensus package that contains the canonical and extended fork definitions.
func (a *AbciApp) Activations(height int64) []*consensus.Hardfork {
	// hmm, this is a tup of the TxApp method
	var activations []*consensus.Hardfork
	activationNames := a.forks.ActivatesAt(uint64(height))
	for _, name := range activationNames {
		fork := consensus.Hardforks[name]
		if fork == nil {
			a.log.Errorf("hardfork %v at height %d has no definition", name, height)
			continue // really could be a panic
		}
		activations = append(activations, fork)
	}
	return activations
}

var _ abciTypes.Application = &AbciApp{}

// The Application interface methods in four groups according to the
// "connection" used by CometBFT to interact with the application. Calls to the
// methods within a connection are synchronized. They are not synchronized
// between the connections. e.g. CheckTx calls from the mempool connection can
// occur concurrent to calls on the Consensus connection.

// CheckTx is the "Guardian of the mempool: every node runs CheckTx before
// letting a transaction into its local mempool". Also "The transaction may come
// from an external user or another node". Further "CheckTx validates the
// transaction against the current state of the application, for example,
// checking signatures and account balances, but does not apply any of the state
// changes described in the transaction."
//
// This method must reject transactions that are invalid and/or may be crafted
// to attack the network by flooding the mempool or filling blocks with rejected
// transactions.
//
// This method is also used to re-check mempool transactions after blocks are
// mined. This is used to *evict* previously accepted transactions that become
// invalid, which may happen for a variety of reason only the application can
// decide, such as changes in account balance and last mined nonce.
//
// It is important to use this method rather than include failing transactions
// in blocks, particularly if the failure mode involves the transaction author
// spending no gas or achieving including in the block with little effort.
func (a *AbciApp) CheckTx(ctx context.Context, incoming *abciTypes.RequestCheckTx) (*abciTypes.ResponseCheckTx, error) {
	newTx := incoming.Type == abciTypes.CheckTxType_New
	logger := a.log.With(zap.Bool("recheck", !newTx))

	// If the network is halted for migration, we reject all transactions.
	if a.halted.Load() || a.chainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return &abciTypes.ResponseCheckTx{Code: codeInvalidTxType.Uint32(), Log: "network is halted for migration"}, nil
	}

	var err error
	code := codeOk

	// NOTE about the error logging here: These transactions are from users, so
	// most of these are not server errors, but client errors, so we ideally do
	// not want to log them at all in production. We'll keep a few for now to
	// help debugging.

	tx := &transactions.Transaction{}
	err = tx.UnmarshalBinary(incoming.Tx)
	if err != nil {
		code = codeEncodingError
		logger.Debug("failed to unmarshal transaction", zap.Error(err))
		return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil // return error now or is it still all about code?
	}

	logger.Debug("check tx",
		zap.String("sender", hex.EncodeToString(tx.Sender)),
		zap.String("PayloadType", tx.Body.PayloadType.String()),
		zap.Uint64("nonce", tx.Body.Nonce))

	// For a new transaction (not re-check), before looking at execution cost or
	// checking nonce validity, ensure the payload is recognized and signature is valid.
	if newTx {
		// Verify the correct chain ID is set, if it is set.
		if protected := tx.Body.ChainID != ""; protected && tx.Body.ChainID != a.cfg.ChainID {
			code = codeWrongChain
			logger.Info("wrong chain ID",
				zap.String("payloadType", tx.Body.PayloadType.String()))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: "wrong chain ID"}, nil
		}

		// Verify Signature
		err = ident.VerifyTransaction(tx)
		if err != nil {
			code = codeInvalidSignature
			logger.Debug("failed to verify transaction", zap.Error(err))
			return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil
		}
	} else {
		logger.Debug("Recheck", zap.String("sender", hex.EncodeToString(tx.Sender)), zap.Uint64("nonce", tx.Body.Nonce), zap.String("payloadType", tx.Body.PayloadType.String()))
	}

	readTx, err := a.db.BeginReadTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin read tx: %w", err)
	}
	defer readTx.Rollback(ctx) // always rollback since we are read-only

	auth, err := authExt.GetAuthenticator(tx.Signature.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticator: %w", err)
	}

	ident, err := auth.Identifier(tx.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get identifier: %w", err)
	}

	err = a.txApp.ApplyMempool(&common.TxContext{
		Ctx: ctx,
		BlockContext: &common.BlockContext{
			ChainContext: a.chainContext,
			Height:       a.height + 1, // height increments at the start of FinalizeBlock,
			Proposer:     nil,          // we don't know the proposer here
		},
		TxID:          hex.EncodeToString(cometTXID(incoming.Tx)),
		Signer:        tx.Sender,
		Caller:        ident,
		Authenticator: tx.Signature.Type,
	}, readTx, tx)
	if err != nil {
		if errors.Is(err, transactions.ErrInvalidNonce) {
			code = codeInvalidNonce
			logger.Info("received transaction with invalid nonce", zap.Uint64("nonce", tx.Body.Nonce), zap.Error(err), zap.String("payloadType", tx.Body.PayloadType.String()))
		} else if errors.Is(err, transactions.ErrInvalidAmount) {
			code = codeInvalidAmount
			logger.Info("received transaction with invalid amount", zap.Uint64("nonce", tx.Body.Nonce), zap.Error(err))
		} else if errors.Is(err, transactions.ErrInsufficientBalance) {
			code = codeInsufficientBalance
			logger.Info("transaction sender has insufficient balance", zap.Uint64("nonce", tx.Body.Nonce), zap.Error(err))
		} else {
			code = codeUnknownError
			logger.Warn("unexpected failure to verify transaction against local mempool state", zap.Error(err))
		}
		// Evicting transaction from mempool.
		txHash := sha256.Sum256(incoming.Tx)
		a.verifiedTxnsMtx.Lock()
		delete(a.verifiedTxns, txHash)
		a.verifiedTxnsMtx.Unlock()
		return &abciTypes.ResponseCheckTx{Code: code.Uint32(), Log: err.Error()}, nil
	}

	// Cache the transaction hash
	if newTx {
		txHash := sha256.Sum256(incoming.Tx)
		a.verifiedTxnsMtx.Lock()
		a.verifiedTxns[txHash] = struct{}{}
		a.verifiedTxnsMtx.Unlock()
	}
	return &abciTypes.ResponseCheckTx{Code: code.Uint32()}, nil
}

// cometTXID gets the cometbft transaction ID.
func cometTXID(tx []byte) []byte {
	return tmhash.Sum(tx)
}

// FinalizeBlock is on the consensus connection. Note that according to CometBFT
// docs, "ResponseFinalizeBlock.app_hash is included as the Header.AppHash in
// the next block."
func (a *AbciApp) FinalizeBlock(ctx context.Context, req *abciTypes.RequestFinalizeBlock) (*abciTypes.ResponseFinalizeBlock, error) {
	logger := a.log.With(zap.String("stage", "ABCI FinalizeBlock"), zap.Int64("height", req.Height))

	if a.genesisTx != nil {
		// if we are in the first block, we use the genesisTx.
		// This is to prevent a bug where a node crashing between InitChain and
		// genesis is unable to recover.
		a.consensusTx = a.genesisTx
		a.genesisTx = nil
	} else {
		var err error
		a.consensusTx, err = a.db.BeginPreparedTx(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin outer tx failed: %w", err)
		}
	}

	err := a.txApp.Begin(ctx, req.Height)
	if err != nil {
		return nil, fmt.Errorf("begin tx commit failed: %w", err)
	}

	a.chainContextMtx.Lock()
	defer a.chainContextMtx.Unlock()

	// we copy the Kwil consensus params to ensure we persist any changes
	// made during the block execution
	networkParams := &common.NetworkParameters{
		MaxBlockSize:     a.consensusParams.Block.MaxBytes,
		JoinExpiry:       a.consensusParams.Validator.JoinExpiry,
		VoteExpiry:       a.consensusParams.Votes.VoteExpiry,
		DisabledGasCosts: a.consensusParams.WithoutGasCosts,
		MaxVotesPerTx:    a.consensusParams.Votes.MaxVotesPerTx,
		MigrationStatus:  a.chainContext.NetworkParameters.MigrationStatus,
	}
	oldNetworkParams := *networkParams

	initialValidators, err := a.txApp.GetValidators(ctx, a.consensusTx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current validators: %w", err)
	}

	// Punish bad validators.
	for _, ev := range req.Misbehavior {
		addr := proposerAddrToString(ev.Validator.Address) // comet example app confirms this conversion... weird
		logger.Info("punish validator", zap.String("addr", addr))
		// FORKSITE: could alter punishment system (consider misbehavior Type)

		// CometBFT gives the address, not public key, so we have to remember them.
		pubkey, ok := a.validatorAddressToPubKey[addr]
		if !ok {
			// It is possible or likely that a misbehaving validator will
			// misbehave in consecutive blocks, so it should not be an error
			// here since it could merely be a subsequent misbehavior if we have
			// removed them from our app's map.
			continue
		}

		const punishDelta = 1
		newPower := ev.Validator.Power - punishDelta
		if err = a.txApp.UpdateValidator(ctx, a.consensusTx, pubkey, newPower); err != nil {
			return nil, fmt.Errorf("failed to punish validator: %w", err)
		}
	}

	addr := proposerAddrToString(req.ProposerAddress)
	proposerPubKey, ok := a.validatorAddressToPubKey[addr]
	if !ok && len(req.Txs) > 0 {
		// ProcessProposal allows block proposals for untracked validators, but
		// only if the block has no transactions.
		return nil, fmt.Errorf("failed to find proposer pubkey corresponding to address %v", addr)
	}
	// Note that in the !ok case, empty Txs is required, and the proposerPubKey
	// may be empty!

	res := &abciTypes.ResponseFinalizeBlock{}

	blockCtx := common.BlockContext{
		ChainContext: a.chainContext,
		Height:       req.Height,
		Timestamp:    req.Time.Unix(),
		Proposer:     proposerPubKey,
	}

	inMigration := blockCtx.ChainContext.NetworkParameters.MigrationStatus == types.MigrationInProgress
	haltNetwork := blockCtx.ChainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted

	// since notifications are returned async from postgres, we will construct
	// a map to track them in, and wait at the end of the function to add them
	// to the ResponseFinalizeBlock

	// maps the transaction hash to the logs for that transaction
	type logTracker struct {
		logs      string
		truncated bool
	}
	logMap := make(map[string]*logTracker)
	// resultArr tracks the txHash and the abci result for each tx.
	// This is necessary to avoid recomputing the hash for all txs
	type txResult struct {
		TxHash []byte
		Result *abciTypes.ExecTxResult
	}
	resultArr := make([]*txResult, len(req.Txs))

	// subscribe to any notifications
	logs, done, err := a.consensusTx.Subscribe(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to notifications: %w", err)
	}
	defer done(ctx)

	// wait group to wait at the end of the function for all logs to be received
	logsDone := make(chan error, 1)
	go func() {
		defer close(logsDone)

		// we enforce that the cumulative size of logs is less than 1KB
		// per tx. This is a work-around until we have gas costs to protect
		// against log spam.
		for {
			log, ok := <-logs
			if !ok { // empty and closed (normal completion)
				return // logsDone receiver gets nil error
			}
			if log == "" {
				// The DB has shut down with active subscribers. Fail.
				logsDone <- errors.New("premature notice stream termination")
				return
			}
			txid, notice, err := parseCommon.ParseNotice(log)
			if err != nil {
				// will still be deterministic so nbd to not halt here
				a.log.Errorf("failed to parse notice (%.20s...): %v", log, err)
				continue // since txid is invalid and won't match any result.TxHash
			}

			currentLog, ok := logMap[txid]
			if !ok {
				logMap[txid] = &logTracker{
					logs:      "",
					truncated: false,
				}
				currentLog = logMap[txid]
			}
			if len(currentLog.logs)+len(notice) > 1024 {
				if !currentLog.truncated {
					currentLog.logs += "\n[truncated]"
					currentLog.truncated = true
				}
				// else, we have already truncated the log
			} else {
				currentLog.logs += "\n" + notice
			}
		}
	}()

	for i, tx := range req.Txs {
		decoded := &transactions.Transaction{}
		err := decoded.UnmarshalBinary(tx)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
		}

		txHash := sha256.Sum256(tx)

		auth, err := authExt.GetAuthenticator(decoded.Signature.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get authenticator: %w", err)
		}

		ident, err := auth.Identifier(decoded.Sender)
		if err != nil {
			return nil, fmt.Errorf("failed to get identifier: %w", err)
		}

		txRes := a.txApp.Execute(&common.TxContext{
			Ctx:           ctx,
			TxID:          hex.EncodeToString(txHash[:]), // tmhash.Sum(tx), // use cometbft TmHash to get the same hash as is indexed
			BlockContext:  &blockCtx,
			Signer:        decoded.Sender,
			Authenticator: decoded.Signature.Type,
			Caller:        ident,
		}, a.consensusTx, decoded)

		abciRes := &abciTypes.ExecTxResult{}
		if txRes.Error != nil {
			abciRes.Log = txRes.Error.Error()
			a.log.Debug("failed to execute transaction", zap.Error(txRes.Error))
		} else {
			abciRes.Log = "success"
		}
		abciRes.Code = txRes.ResponseCode.Uint32()
		abciRes.GasUsed = txRes.Spend

		resultArr[i] = &txResult{
			TxHash: txHash[:],
			Result: abciRes,
		}

		res.TxResults = append(res.TxResults, abciRes)

		// Remove the transaction from the cache as it has been included in a block
		a.verifiedTxnsMtx.Lock()
		delete(a.verifiedTxns, txHash)
		a.verifiedTxnsMtx.Unlock()
	}

	// If at activation height, submit any consensus params updates associated
	// with the fork. They should not overlap (some forks should be superseded
	// by later fork definitions and not used on new networks).
	paramUpdates := consensus.ParamUpdates{}
	for _, fork := range a.Activations(req.Height) {
		if fork.ParamsUpdates == nil {
			continue
		}
		consensus.MergeConsensusUpdates(&paramUpdates, fork.ParamsUpdates)
	}

	// Merge, including kwil-specific params like join expiry.
	updateConsensusParams(a.consensusParams, &paramUpdates)

	// merge the Kwil-specific params into the network params
	networkParams.MaxBlockSize = a.consensusParams.Block.MaxBytes
	networkParams.JoinExpiry = a.consensusParams.Validator.JoinExpiry
	networkParams.VoteExpiry = a.consensusParams.Votes.VoteExpiry
	networkParams.MaxVotesPerTx = a.consensusParams.Votes.MaxVotesPerTx
	networkParams.DisabledGasCosts = a.consensusParams.WithoutGasCosts

	// cometbft wants its api/tendermint type
	res.ConsensusParamUpdates = cometbft.ParamUpdatesToComet(&paramUpdates)

	// Broadcast any events that have not been broadcasted yet
	if a.broadcastFn != nil && len(proposerPubKey) > 0 {
		err := a.broadcastFn(ctx, a.consensusTx, &blockCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast events: %w", err)
		}
	}

	a.log.Debug("Finalize(start)", log.Int("height", a.height), log.String("appHash", hex.EncodeToString(a.appHash)))
	// Get the new validator set and apphash from txApp.
	finalValidators, approvedJoins, expiredJoins, err := a.txApp.Finalize(ctx, a.consensusTx, &blockCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize transaction app: %w", err)
	}

	// Notify the migrator of the changeset
	err = a.migrator.NotifyHeight(ctx, &blockCtx, a.db)
	if err != nil {
		return nil, fmt.Errorf("failed to notify migrator of changeset: %w", err)
	}

	networkParams.MigrationStatus = a.chainContext.NetworkParameters.MigrationStatus

	// store any changes to the network params
	err = meta.StoreDiff(ctx, a.consensusTx, &oldNetworkParams, networkParams)
	if err != nil {
		return nil, fmt.Errorf("failed to store network params diff: %w", err)
	}

	// While still in the DB transaction, update to this next height but dummy
	// app hash. If we crash before Commit can store the next app hash that we
	// get after Precommit, the startup handshake's call to Info will detect the
	// mismatch. That requires manual recovery (drop state and reapply), but it
	// at least detects this recorded height rather than not recognizing that we
	// have committed the data for this block at all.
	err = meta.SetChainState(ctx, a.consensusTx, req.Height, []byte{0x42})
	if err != nil {
		return nil, err
	}

	// Create a new changeset processor
	csp := newChangesetProcessor()
	// "migrator" module subscribes to the changeset processor to store changesets during the migration
	csErrChan := make(chan error, 1)
	defer close(csErrChan)

	if inMigration && !haltNetwork {
		csChanMigrator, err := csp.Subscribe(ctx, "migrator")
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe to changeset processor: %w", err)
		}
		// migrator go routine will receive changesets from the changeset processor
		// give the new channel to the migrator to store changesets
		go func() {
			csErrChan <- a.migrator.StoreChangesets(req.Height, csChanMigrator)
		}()
	}

	// statistics module can subscribe to the changeset processor to listen for changesets for updating statistics
	// statsChan := csp.Subscribe(ctx, "statistics")

	// changeset processor is not ready to receive changesets and  broadcast them to subscribers
	go csp.BroadcastChangesets(ctx)

	// we now get the apphash by calling precommit on the transaction
	appHash, err := a.consensusTx.Precommit(ctx, csp.csChan)
	if err != nil {
		return nil, fmt.Errorf("failed to precommit transaction: %w", err)
	}

	err = done(ctx) // inform pg that no more logs are needed
	if err != nil {
		return nil, fmt.Errorf("failed to close subscription: %w", err)
	}

	a.stateMtx.Lock()
	newAppHash := sha256.Sum256(append(a.appHash, appHash...))
	res.AppHash = newAppHash[:]
	a.appHash = newAppHash[:]
	a.height = req.Height
	a.stateMtx.Unlock()

	if a.forks.BeginsHalt(uint64(req.Height) - 1) {
		a.log.Info("This is the last block before halt.")
		a.halted.Store(true)
	}

	valUpdates := validatorUpdates(initialValidators, finalValidators)

	res.ValidatorUpdates = make([]abciTypes.ValidatorUpdate, len(valUpdates))
	for i, up := range valUpdates {
		addr, err := cometbft.PubkeyToAddr(up.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
		}
		if up.Power == 0 {
			delete(a.validatorAddressToPubKey, addr)
			if err = a.p2p.RemovePeer(ctx, addr); err != nil {
				if !errors.Is(err, cometbft.ErrPeerNotWhitelisted) {
					a.log.Warn("failed to remove demoted validator from peer list", log.String("address", addr), log.Error(err))
				}
			}
		} else {
			a.validatorAddressToPubKey[addr] = up.PubKey // there may be new validators we need to add
			// Add the validator to the peer list
			if err = a.p2p.AddPeer(ctx, addr); err != nil {
				if !errors.Is(err, cometbft.ErrPeerAlreadyWhitelisted) {
					a.log.Warn("failed to whitelist promoted validator", log.String("address", addr), log.Error(err))
				}
			}
		}

		res.ValidatorUpdates[i] = abciTypes.Ed25519ValidatorUpdate(up.PubKey, up.Power)
	}

	// Join requests approved by this node are added to the peer list.
	for _, pubKey := range approvedJoins {
		addr, err := cometbft.PubkeyToAddr(pubKey)
		if err != nil {
			a.log.Warn("failed to convert pubkey to address", log.Error(err))
			continue
		}
		if err = a.p2p.AddPeer(ctx, addr); err != nil {
			if !errors.Is(err, cometbft.ErrPeerAlreadyWhitelisted) {
				a.log.Warn("failed to whitelist new validator", log.String("address", addr), log.Error(err))
			}
		}
	}

	// peers whose join requests have expired are removed from the peer list
	for _, pubKey := range expiredJoins {
		addr, err := cometbft.PubkeyToAddr(pubKey)
		if err != nil {
			a.log.Warn("failed to convert pubkey to address", zap.Error(err))
		}
		if err = a.p2p.RemovePeer(ctx, addr); err != nil {
			if !errors.Is(err, cometbft.ErrPeerNotWhitelisted) {
				a.log.Warn("failed to remove expired validator from peer list", log.String("address", addr), log.Error(err))
			}
		}
	}

	// wait for all logs to be received, or a premature shutdown
	select {
	case <-ctx.Done():
		// NOTE: this will not happen until cometbft v1.0 since in v0.38
		// FinalizeBlock is still called with context.TODO().
		return nil, ctx.Err()
	case err := <-logsDone:
		if err != nil {
			return nil, fmt.Errorf("DB failure: %w", err)
		} // else we got all the notice
	}

	for _, result := range resultArr {
		logs, ok := logMap[hex.EncodeToString(result.TxHash)]
		if !ok {
			continue
		}
		result.Result.Log += logs.logs
	}

	if inMigration && !haltNetwork {
		// wait for the migrator to finish storing changesets
		err = <-csErrChan
		if err != nil {
			return nil, fmt.Errorf("failed to store changesets: %w", err)
		}
	}

	// Persist app hash and height to the disk for recovery purposes.
	lc := &lastCommitInfo{
		Height:  req.Height,
		AppHash: newAppHash[:],
	}
	if err = lc.saveAs(a.lastCommitInfoFileName); err != nil {
		a.log.Warn("failed to save last commit info", log.Error(err))
	}

	return res, nil
}

// validatorUpdates returns the added, removed, and updated validators in the given block.
func validatorUpdates(initial, final []*types.Validator) []*types.Validator {
	initialVals := make(map[string]*types.Validator)
	for _, val := range initial {
		initialVals[hex.EncodeToString(val.PubKey)] = val
	}

	finalVals := make(map[string]*types.Validator)
	for _, val := range final {
		finalVals[hex.EncodeToString(val.PubKey)] = val
	}

	var updates []*types.Validator
	// check for newly added and updated validators
	for key, val := range finalVals {
		if initialVal, ok := initialVals[key]; ok {
			if initialVal.Power != val.Power {
				// Validator is in the initial set, but has updated power
				updates = append(updates, val)
			}
		} else {
			// Validator is not in the initial set, so it must be new
			updates = append(updates, val)
		}
	}

	// check for removed validators
	for key, val := range initialVals {
		if _, ok := finalVals[key]; !ok {
			// Validator is in the initial set, but not in the final set
			updates = append(updates, &types.Validator{
				PubKey: val.PubKey,
				Power:  0,
			})
		}
	}

	return updates
}

// Commit persists the state changes. This is called under mempool lock in
// cometbft, unlike FinalizeBlock.
func (a *AbciApp) Commit(ctx context.Context, _ *abciTypes.RequestCommit) (*abciTypes.ResponseCommit, error) {
	err := a.consensusTx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction app: %w", err)
	}

	// we need to re-open a new transaction just to write the apphash
	// TODO: it would be great to have a way to commit the apphash without
	// opening a new transaction. This could leave us in a state where data is
	// committed but the apphash is not, which would essentially nuke the chain.
	ctx0 := context.Background() // badly timed shutdown MUST NOT cancel now, we need consistency with consensus tx commit
	tx, err := a.db.BeginTx(ctx0)
	if err != nil {
		return nil, fmt.Errorf("failed to begin outer tx: %w", err)
	}

	a.stateMtx.Lock()
	height, appHash := a.height, a.appHash
	a.stateMtx.Unlock()
	err = meta.SetChainState(ctx0, tx, height, appHash)
	if err != nil {
		err2 := tx.Rollback(ctx0)
		if err2 != nil {
			return nil, fmt.Errorf("failed to rollback transaction app: %w", err2)
		}
		return nil, fmt.Errorf("failed to set chain state: %w", err)
	}

	err = a.migrator.PersistLastChangesetHeight(ctx0, tx)
	if err != nil {
		err2 := tx.Rollback(ctx0)
		if err2 != nil {
			return nil, fmt.Errorf("failed to rollback transaction app: %w", err2)
		}
		return nil, fmt.Errorf("failed to persist last changeset height: %w", err)
	}

	err = tx.Commit(ctx0)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction app: %w", err)
	}

	a.txApp.Commit(ctx)

	// Snapshots are to be taken if:
	// - the block height is a multiple of the snapshot interval
	// - there are no snapshots in the store (This is to support the new nodes joining the network using
	//   statesync. As the snapshot intervals are pretty large, the new nodes will not have any
	//   snapshots to start with for a long time, during which the new nodes just hangs in the snapshot
	//   discovery phase, which can be mitigated if we can produce a snapshot at the start of the network
	//   or when the node joins network at any height)
	snapshotsDue := a.snapshotter != nil &&
		(a.snapshotter.IsSnapshotDue(uint64(a.height)) || len(a.snapshotter.ListSnapshots()) == 0)
	snapshotsDue = snapshotsDue && a.height > max(1, a.cfg.InitialHeight)

	if a.replayingBlocks != nil && snapshotsDue && !a.replayingBlocks() {
		// we make a snapshot tx but don't directly use it. This is because under the hood,
		// we are using the pg_dump executable to create the snapshot, and we are simply
		// giving pg_dump the snapshot ID to guarantee it has an isolated view of the database.
		snapshotTx, snapshotId, err := a.db.BeginSnapshotTx(ctx)
		if err != nil {
			a.log.Error("failed to begin snapshot tx", zap.Error(err))
			return &abciTypes.ResponseCommit{}, nil
		}
		defer snapshotTx.Rollback(ctx) // always rollback, since this is just for view isolation

		err = a.snapshotter.CreateSnapshot(ctx, uint64(a.height), snapshotId, statesyncSnapshotSchemas, statsyncExcludedTables, nil)
		if err != nil {
			a.log.Error("failed to create snapshot", zap.Error(err))
		} else {
			a.log.Info("created snapshot", zap.Uint64("height", uint64(a.height)), zap.String("snapshot_id", snapshotId))
		}
	}

	// If a broadcast was accepted during execution of that block, it will be
	// rechecked and possibly evicted immediately following our commit Response.

	return &abciTypes.ResponseCommit{}, nil // RetainHeight stays 0 to not prune any blocks
}

// lastCommitInfo is a struct to store the last commit info such as
// height and apphash at the end of the FinalizeBlock method.
type lastCommitInfo struct {
	AppHash types.HexBytes `json:"app_hash"`
	Height  int64          `json:"height"`
}

// saveAs saves the last commit info to the given file.
func (lc *lastCommitInfo) saveAs(filename string) error {
	bts, err := json.MarshalIndent(lc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal last commit info: %w", err)
	}

	return os.WriteFile(filename, bts, 0644)
}

// loadLastCommitInfo loads the last commit info from the given file.
func loadLastCommitInfo(filename string) (*lastCommitInfo, error) {
	bts, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read last commit info: %w", err)
	}

	var lc lastCommitInfo
	if err = json.Unmarshal(bts, &lc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal last commit info: %w", err)
	}

	return &lc, nil
}

// Info is part of the Info/Query connection. CometBFT will call this during
// it's handshake, and if height 0 is returned it will then use InitChain. This
// method should also be usable at any time (and read only) as it is used for
// cometbft's /abci_info RPC endpoint.
//
// CometBFT docs say:
//   - LastBlockHeight is the "Latest height for which the app persisted its state"
//   - LastBlockAppHash is the "Latest AppHash returned by FinalizeBlock"
//   - "ResponseFinalizeBlock.app_hash is included as the Header.AppHash in the
//     next block." Notably, the *next* block's header. This is verifiable with
//     the /block RPC endpoint.
//
// Thus, LastBlockAppHash is not NOT AppHash in the header of the block at the
// returned height. Our meta data store has to track the above values, where the
// stored app hash corresponds to the block at height+1. This is simple, but the
// discrepancy is worth noting.
func (a *AbciApp) Info(ctx context.Context, _ *abciTypes.RequestInfo) (*abciTypes.ResponseInfo, error) {
	a.stateMtx.Lock()
	if a.height > 0 { // has already been set and stored in FinalizeBlock
		defer a.stateMtx.Unlock()
		return &abciTypes.ResponseInfo{
			LastBlockHeight:  a.height,
			LastBlockAppHash: a.appHash,
			Version:          version.KwilVersion, // the *software* semver string
			AppVersion:       a.cfg.ApplicationVersion,
		}, nil
	}
	a.stateMtx.Unlock()
	// else we're probably responding to the ABCI "handshake" and need to read
	// chain state from app DB.

	readTx, err := a.db.BeginReadTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin read tx: %w", err)
	}
	defer readTx.Rollback(ctx) // always rollback since we are read-only

	height, appHash, err := meta.GetChainState(ctx, readTx)
	if err != nil {
		return nil, fmt.Errorf("GetChainState: %w", err)
	}
	if height == -1 {
		height = 0 // for ChainInfo caller, non-negative is expected for genesis
	}

	a.log.Info("ABCI application is ready", zap.Int64("height", height))

	return &abciTypes.ResponseInfo{
		LastBlockHeight:  height,
		LastBlockAppHash: appHash,
		Version:          version.KwilVersion, // the *software* semver string
		AppVersion:       a.cfg.ApplicationVersion,
	}, nil
}

func (a *AbciApp) InitChain(ctx context.Context, req *abciTypes.RequestInitChain) (*abciTypes.ResponseInitChain, error) {
	logger := a.log.With(zap.String("stage", "ABCI InitChain"), zap.Int64("height", req.InitialHeight))
	logger.Debug("", zap.String("ChainId", req.ChainId))
	// maybe verify a.cfg.ChainID against the one in the request
	var err error
	a.genesisTx, err = a.db.BeginPreparedTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin outer tx failed: %w", err)
	}

	// Store the genesis account allocations to the datastore. These are
	// reflected in the genesis app hash.
	genesisAllocs := make([]*types.Account, 0, len(a.cfg.GenesisAllocs))
	for acct, bal := range a.cfg.GenesisAllocs {
		acct, _ := strings.CutPrefix(acct, "0x") // special case for ethereum addresses
		identifier, err := hex.DecodeString(acct)
		if err != nil {
			return nil, fmt.Errorf("invalid hex pubkey: %w", err)
		}

		genesisAllocs = append(genesisAllocs, &types.Account{
			Identifier: identifier,
			Balance:    bal,
		})
	}
	// Initialize the validator module with the genesis validators.
	vldtrs := make([]*types.Validator, len(req.Validators))
	for i := range req.Validators {
		vi := &req.Validators[i]
		pk := vi.PubKey.GetEd25519()
		vldtrs[i] = &types.Validator{
			PubKey: pk,
			Power:  vi.Power,
		}

		addr, err := cometbft.PubkeyToAddr(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
		}

		a.validatorAddressToPubKey[addr] = pk
	}

	// With the genesisTx not being committed until the first FinalizeBlock, we
	// expect no existing chain state in the application (postgres).
	height, appHash, err := meta.GetChainState(ctx, a.genesisTx)
	if err != nil {
		return nil, fmt.Errorf("error getting database height: %s", err.Error())
	}

	// First app hash and height are stored in FinalizeBlock for first block.
	if height != -1 {
		return nil, fmt.Errorf("expected database to be uninitialized, but had height %d", height)
	}
	if len(appHash) != 0 {
		return nil, fmt.Errorf("expected NULL app hash, got %x", appHash)
	}

	startParams := *a.chainContext.NetworkParameters

	if err := a.txApp.GenesisInit(ctx, a.genesisTx, vldtrs, genesisAllocs, req.InitialHeight, a.chainContext); err != nil {
		return nil, fmt.Errorf("txApp.GenesisInit failed: %w", err)
	}

	// persist any diff to the network params
	err = meta.StoreDiff(ctx, a.genesisTx, &startParams, a.chainContext.NetworkParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to store network params diff: %w", err)
	}

	valUpdates := make([]abciTypes.ValidatorUpdate, len(vldtrs))
	for i, validator := range vldtrs {
		valUpdates[i] = abciTypes.Ed25519ValidatorUpdate(validator.PubKey, validator.Power)
	}

	logger.Info("initialized chain", zap.String("app hash", fmt.Sprintf("%x", a.cfg.GenesisAppHash)))

	return &abciTypes.ResponseInitChain{
		Validators: valUpdates,
		AppHash:    a.cfg.GenesisAppHash, // doesn't matter what we gave the node in GenesisDoc, this is it
	}, nil
}

// ApplySnapshotChunk is on the state sync connection
func (a *AbciApp) ApplySnapshotChunk(ctx context.Context, req *abciTypes.RequestApplySnapshotChunk) (*abciTypes.ResponseApplySnapshotChunk, error) {
	if a.statesyncer == nil {
		return nil, fmt.Errorf("mismatched statesync configuration between CometBFT and ABCI app")
	}

	dbRestored, err := a.statesyncer.ApplySnapshotChunk(ctx, req.Chunk, req.Index)
	if err != nil {
		var refetchChunks []uint32
		// If the chunk was not applied successfully either due to chunk hash mismatch or other reasons,
		// refetch the chunk from other peers
		if errors.Is(err, statesync.ErrRefetchSnapshotChunk) {
			refetchChunks = append(refetchChunks, req.Index)
		}
		a.log.Errorf("Failed to apply snapshot chunk: %v", err)
		return &abciTypes.ResponseApplySnapshotChunk{
			Result:        statesync.ToABCIApplySnapshotChunkResponse(err),
			RefetchChunks: refetchChunks,
		}, nil
	}

	if dbRestored {
		readTx, err := a.db.BeginReadTx(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to begin read tx: %w", err)
		}
		defer readTx.Rollback(ctx) // always rollback since we are read-only

		// DB restored successfully from the snapshot
		// Update the engine's in-memory info with the new database state
		err = a.txApp.Reload(ctx, readTx)
		if err != nil {
			return &abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ABORT}, err
		}

		// Cache the pubkey in the validatorAddressToPubKey map, as the map is not previously populated
		validators, err := a.txApp.GetValidators(ctx, readTx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current validators: %w", err)
		}
		for _, val := range validators {
			addr, err := cometbft.PubkeyToAddr(val.PubKey)
			if err != nil {
				return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
			}

			a.validatorAddressToPubKey[addr] = val.PubKey
		}

		// Update the app Hash
		height, appHash, err := meta.GetChainState(ctx, readTx)
		if err != nil {
			return nil, fmt.Errorf("GetChainState: %w", err)
		}

		a.stateMtx.Lock()
		a.appHash = appHash
		a.height = height
		a.stateMtx.Unlock()
	}

	return &abciTypes.ResponseApplySnapshotChunk{Result: abciTypes.ResponseApplySnapshotChunk_ACCEPT, RefetchChunks: nil}, nil
}

// ListSnapshots is on the state sync connection
func (a *AbciApp) ListSnapshots(ctx context.Context, req *abciTypes.RequestListSnapshots) (*abciTypes.ResponseListSnapshots, error) {
	if a.snapshotter == nil {
		return &abciTypes.ResponseListSnapshots{}, nil
	}

	snapshots := a.snapshotter.ListSnapshots()

	var res []*abciTypes.Snapshot
	for _, snapshot := range snapshots {
		bts, err := json.Marshal(snapshot)
		if err != nil {
			return nil, err
		}

		sp := &abciTypes.Snapshot{
			Height:   snapshot.Height,
			Format:   snapshot.Format,
			Chunks:   snapshot.ChunkCount,
			Hash:     snapshot.SnapshotHash,
			Metadata: make([]byte, len(bts)),
		}
		copy(sp.Metadata, bts)

		res = append(res, sp)
	}
	return &abciTypes.ResponseListSnapshots{Snapshots: res}, nil
}

// LoadSnapshotChunk is on the state sync connection
func (a *AbciApp) LoadSnapshotChunk(ctx context.Context, req *abciTypes.RequestLoadSnapshotChunk) (*abciTypes.ResponseLoadSnapshotChunk, error) {
	if a.snapshotter == nil {
		return &abciTypes.ResponseLoadSnapshotChunk{}, nil
	}

	chunk, err := a.snapshotter.LoadSnapshotChunk(req.Height, req.Format, req.Chunk)
	if err != nil {
		return nil, err
	}
	return &abciTypes.ResponseLoadSnapshotChunk{Chunk: chunk}, nil
}

// OfferSnapshot is on the state sync connection
func (a *AbciApp) OfferSnapshot(ctx context.Context, req *abciTypes.RequestOfferSnapshot) (*abciTypes.ResponseOfferSnapshot, error) {
	if a.statesyncer == nil {
		return &abciTypes.ResponseOfferSnapshot{
				Result: abciTypes.ResponseOfferSnapshot_REJECT},
			fmt.Errorf("mismatched statesync configuration between CometBFT and ABCI app")
	}

	var snapshot statesync.Snapshot
	err := json.Unmarshal(req.Snapshot.Metadata, &snapshot)
	if err != nil {
		return &abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_REJECT}, err
	}

	err = a.statesyncer.OfferSnapshot(ctx, &snapshot)
	if err != nil {
		return &abciTypes.ResponseOfferSnapshot{Result: statesync.ToABCIOfferSnapshotResponse(err)}, nil
	}
	return &abciTypes.ResponseOfferSnapshot{Result: abciTypes.ResponseOfferSnapshot_ACCEPT}, nil
}

// ExtendVote creates an application specific vote extension.
//
//   - ResponseExtendVote.vote_extension is application-generated information that
//     will be signed by CometBFT and attached to the Precommit message.
//   - The Application may choose to use an empty vote extension (0 length).
//   - The contents of RequestExtendVote correspond to the proposed block on which
//     the consensus algorithm will send the Precommit message.
//   - ResponseExtendVote.vote_extension will only be attached to a non-nil
//     Precommit message. If the consensus algorithm is to precommit nil, it will
//     not call RequestExtendVote.
//   - The Application logic that creates the extension can be non-deterministic.
func (a *AbciApp) ExtendVote(ctx context.Context, req *abciTypes.RequestExtendVote) (*abciTypes.ResponseExtendVote, error) {
	return &abciTypes.ResponseExtendVote{}, nil
}

// Verify application's vote extension data
func (a *AbciApp) VerifyVoteExtension(ctx context.Context, req *abciTypes.RequestVerifyVoteExtension) (*abciTypes.ResponseVerifyVoteExtension, error) {
	if len(req.VoteExtension) > 0 {
		// We recognize no vote extensions yet.
		return &abciTypes.ResponseVerifyVoteExtension{
			Status: abciTypes.ResponseVerifyVoteExtension_REJECT,
		}, nil
	}
	return &abciTypes.ResponseVerifyVoteExtension{
		Status: abciTypes.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

// txSubList implements sort.Interface to perform in-place sorting of a slice
// that is a subset of another slice, reordering in both while staying within
// the subsets positions in the parent slice.
//
// For example:
//
//	parent slice: {a0, b2, b0, a1, b1}
//	b's subset: {b2, b0, b1}
//	sorted subset: {b0, b1, b2}
//	parent slice: {a0, b0, b1, a1, b2}
//
// The set if locations used by b elements within the parent slice is unchanged,
// but the elements are sorted.
type txSubList struct {
	sub   []*indexedTxn // sort.Sort references only this with Len and Less
	super []*indexedTxn // sort.Sort also Swaps in super using the i field
}

func (txl txSubList) Len() int {
	return len(txl.sub)
}

func (txl txSubList) Less(i int, j int) bool {
	a, b := txl.sub[i], txl.sub[j]
	return a.Body.Nonce < b.Body.Nonce
}

func (txl txSubList) Swap(i int, j int) {
	// Swap elements in sub.
	txl.sub[i], txl.sub[j] = txl.sub[j], txl.sub[i]
	// Swap the elements in their positions in super.
	ip, jp := txl.sub[i].i, txl.sub[j].i
	txl.super[ip], txl.super[jp] = txl.super[jp], txl.super[ip]
}

// indexedTxn facilitates in-place sorting of transaction slices that are
// subsets of other larger slices using a txSubList. This is only used within
// prepareMempoolTxns, and is package-level rather than scoped to the function
// because we define methods to implement sort.Interface.
type indexedTxn struct {
	i int // index in superset slice
	*transactions.Transaction

	is int // not used for sorting, only referencing the marshalled txn slice
}

// prepareBlockTransactions prepares the transactions for the block we are proposing.
// The input transactions are from mempool direct from cometbft, and we modify
// the list for our purposes. This includes ensuring transactions from the same
// sender in ascending nonce-order, enforcing the max bytes limit, etc.
// This also includes the proposer's transactions, which are not in the mempool.
// The transaction ordering is as follows:
// MempoolProposerTxns, ProposerInjectedTxns, MempoolTxns by other senders
func (a *AbciApp) prepareBlockTransactions(ctx context.Context, txs [][]byte, log *log.Logger, maxTxBytes int64, proposerAddr []byte, height int64) [][]byte {
	// Unmarshal and index the transactions.
	var okTxns []*indexedTxn
	var i int
	for is, txB := range txs {
		tx := &transactions.Transaction{}
		err := tx.UnmarshalBinary(txB)
		if err != nil {
			log.Error("failed to unmarshal transaction that was previously accepted to mempool", zap.Error(err))
			continue // should not have passed CheckTx to get into our mempool
		}
		okTxns = append(okTxns, &indexedTxn{i, tx, is})
		i++
	}

	// Group by sender and stable sort each group by nonce.
	grouped := make(map[string][]*indexedTxn)
	for _, txn := range okTxns {
		key := string(txn.Sender)
		grouped[key] = append(grouped[key], txn)
	}
	for _, txns := range grouped {
		sort.Stable(txSubList{
			sub:   txns,
			super: okTxns,
		})
	}

	// Grab the bytes rather than re-marshalling.
	nonces := make([]uint64, 0, len(okTxns))
	var propTxs, otherTxns []*indexedTxn
	i = 0
	proposerNonce := uint64(0)

	readTx, err := a.db.BeginReadTx(ctx)
	if err != nil {
		log.Error("failed to begin read tx", zap.Error(err))
		return nil
	}
	defer readTx.Rollback(ctx) // always rollback since we are read-only

	// Enforce nonce ordering and remove transactions from unfunded accounts.
	for _, tx := range okTxns {
		if i > 0 && tx.Body.Nonce == nonces[i-1] && bytes.Equal(tx.Sender, okTxns[i-1].Sender) {
			log.Warn(fmt.Sprintf("Dropping tx with re-used nonce %d from block proposal", tx.Body.Nonce))
			continue // mempool recheck should have removed this
		}

		// Enforce the maxVotesPerTx limit for ValidatorVoteIDs transactions
		if tx.Body.PayloadType == transactions.PayloadTypeValidatorVoteIDs {
			voteIDs := &transactions.ValidatorVoteIDs{}
			if err := voteIDs.UnmarshalBinary(tx.Body.Payload); err != nil {
				log.Warn("Dropping tx: failed to unmarshal ValidatorVoteIDs payload", zap.Error(err))
				continue
			}
			if len(voteIDs.ResolutionIDs) > int(a.consensusParams.Votes.MaxVotesPerTx) {
				log.Warn("Dropping tx: ValidatorVoteIDs payload exceeds max votes per tx limits", zap.Int64("max vote limit", a.consensusParams.Votes.MaxVotesPerTx), zap.Int("votes in tx", len(voteIDs.ResolutionIDs)))
				continue
			}
		}

		// Drop transactions from unfunded accounts in gasEnabled mode
		if a.cfg.GasEnabled {
			balance, nonce, err := a.txApp.AccountInfo(ctx, readTx, tx.Sender, false)
			if err != nil {
				log.Error("failed to get account info", zap.Error(err))
				continue
			}
			if nonce == 0 && balance.Sign() == 0 {
				log.Warn("Dropping tx from unfunded account while preparing the block", zap.String("sender", hex.EncodeToString(tx.Sender)))
				continue
			}
		}

		if bytes.Equal(tx.Sender, proposerAddr) {
			proposerNonce = tx.Body.Nonce + 1
			propTxs = append(propTxs, tx)
		} else {
			// Append the transaction to the final list.
			otherTxns = append(otherTxns, tx)
		}
		nonces = append(nonces, tx.Body.Nonce)
		i++
	}

	// TODO: truncate based on our max block size since we'll have to set
	// ConsensusParams.Block.MaxBytes to -1 so that we get ALL transactions even
	// if it goes beyond max_tx_bytes.  See:
	// https://github.com/cometbft/cometbft/pull/1003
	// https://docs.cometbft.com/v0.38/spec/abci/abci++_methods#prepareproposal
	// https://github.com/cometbft/cometbft/issues/980

	// Enforce block size limits
	// Txs order: MempoolProposerTxns, ProposerInjectedTxns, MempoolTxns
	var finalTxs [][]byte

	for _, tx := range propTxs {
		txBts := txs[tx.is]
		txSize := int64(len(txBts))
		if maxTxBytes < txSize {
			break
		}
		maxTxBytes -= txSize
		finalTxs = append(finalTxs, txBts)
	}

	proposerTxs, err := a.txApp.ProposerTxs(ctx, readTx, proposerNonce, maxTxBytes, &common.BlockContext{
		ChainContext: a.chainContext,
		Height:       height,
		Proposer:     proposerAddr,
	})
	if err != nil {
		log.Error("failed to get proposer transactions", zap.Error(err))
	}

	for _, tx := range proposerTxs {
		txSize := int64(len(tx))
		if maxTxBytes < txSize {
			break
		}
		maxTxBytes -= txSize
		finalTxs = append(finalTxs, tx)
	}

	// senders tracks the sender of transactions that has pushed over the bytes limit for the block.
	// If a sender is in the senders, skip all subsequent transactions from the sender
	// because nonces need to be sequential.
	// Keep checking transactions for other senders that may be smaller and fit in the remaining space.
	senders := make(map[string]bool)

	for _, tx := range otherTxns {
		sender := string(tx.Sender)
		// If we have already added a transaction from this sender, skip it.
		if _, ok := senders[sender]; ok {
			continue
		}

		txSize := int64(len(txs[tx.is]))
		if maxTxBytes < txSize {
			// Ignore the rest of the transactions by this sender
			a.log.Warn("transaction being excluded from block with insufficient remaining space",
				zap.String("sender", hex.EncodeToString(tx.Sender)))
			senders[sender] = true
			break
		}
		maxTxBytes -= txSize
		finalTxs = append(finalTxs, txs[tx.is])
	}

	return finalTxs
}

func (a *AbciApp) PrepareProposal(ctx context.Context, req *abciTypes.RequestPrepareProposal) (*abciTypes.ResponsePrepareProposal, error) {
	logger := a.log.With(zap.String("stage", "ABCI PrepareProposal"),
		zap.Int64("height", req.Height),
		zap.Int("txs", len(req.Txs)))

	if a.forks.IsHalt(uint64(req.Height)) || a.chainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		return &abciTypes.ResponsePrepareProposal{}, nil // No more transactions.
	}

	pubKey, ok := a.validatorAddressToPubKey[proposerAddrToString(req.ProposerAddress)]
	if !ok {
		// there is an edge case where cometbft will allow a node to PrepareProposal
		// even if it is not a validator, if it was a validator in the most recent block
		// we check for this here and simply propose an empty block, since it will be rejected
		a.log.Warn("local node was made proposer, but is no longer a validator")
		return &abciTypes.ResponsePrepareProposal{}, nil
	}

	okTxns := a.prepareBlockTransactions(ctx, req.Txs, &a.log, req.MaxTxBytes, pubKey, req.Height)
	if len(okTxns) != len(req.Txs) {
		logger.Info("PrepareProposal: number of transactions in proposed block has changed!",
			zap.Int("in", len(req.Txs)), zap.Int("out", len(okTxns)))
	} else if logger.L.Level() <= log.DebugLevel { // spare the check if it wouldn't be logged
		for i, tx := range okTxns {
			if !bytes.Equal(tx, req.Txs[i]) { //  not and error, just notable
				logger.Debug("transaction was moved or mutated", zap.Int("idx", i))
			}
		}
	}

	return &abciTypes.ResponsePrepareProposal{
		Txs: okTxns,
	}, nil
}

func (a *AbciApp) validateProposalTransactions(ctx context.Context, txns [][]byte, proposer []byte) error {
	logger := a.log.With(zap.String("stage", "ABCI ProcessProposal"))
	grouped, err := groupTxsBySender(txns)
	if err != nil {
		return fmt.Errorf("failed to group transaction by sender: %w", err)
	}

	readTx, err := a.db.BeginReadTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read tx: %w", err)
	}
	defer readTx.Rollback(ctx) // always rollback since we are read-only

	// ensure there are no gaps in an account's nonce, either from the
	// previous best confirmed or within this block. Our current transaction
	// execution does not update an accounts nonce in state unless it is the
	// next nonce. Delivering transactions to a block in that way cannot happen.
	for sender, txs := range grouped {
		_, nonce, err := a.txApp.AccountInfo(ctx, readTx, []byte(sender), false)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}
		expectedNonce := uint64(nonce) + 1

		for _, txO := range txs {
			tx := txO.tx
			if tx.Body.Nonce != expectedNonce {
				logger.Warn("nonce mismatch", zap.Uint64("txNonce", tx.Body.Nonce),
					zap.Uint64("expectedNonce", expectedNonce), zap.String("nonces", fmt.Sprintf("%v", nonceList(txs))))
				return fmt.Errorf("nonce mismatch, expected: %d tx: %d", expectedNonce, tx.Body.Nonce)
			}
			expectedNonce++

			chainID := tx.Body.ChainID
			if protected := chainID != ""; protected && chainID != a.cfg.ChainID {
				return fmt.Errorf("protected transaction with mismatched chain ID")
			}

			// if it is a vote body payload, then only the proposer can propose it
			// this is a hard consensus rule for block building, and is protected by
			// the mempool. The number of Votes in this transaction must not exceed the
			// maxVotesPerTx limits
			if tx.Body.PayloadType == transactions.PayloadTypeValidatorVoteBodies {
				if !bytes.Equal(proposer, tx.Sender) {
					return fmt.Errorf("only proposer can propose validator vote bodies")
				}

				voteBodies := &transactions.ValidatorVoteBodies{}
				if err := voteBodies.UnmarshalBinary(tx.Body.Payload); err != nil {
					return fmt.Errorf("failed to unmarshal vote bodies: %w", err)
				}

				if len(voteBodies.Events) > int(a.consensusParams.Votes.MaxVotesPerTx) {
					return fmt.Errorf("number of events %d in a votebody transaction exceeds the limit %d", len(voteBodies.Events), a.consensusParams.Votes.MaxVotesPerTx)
				}
			}

			// The number of votes in the ValidatorVoteID payload must not be more than
			// maxVotesPerTx limits
			if tx.Body.PayloadType == transactions.PayloadTypeValidatorVoteIDs {
				voteIDs := &transactions.ValidatorVoteIDs{}
				if err := voteIDs.UnmarshalBinary(tx.Body.Payload); err != nil {
					return fmt.Errorf("failed to unmarshal vote ids: %w", err)
				}
				if len(voteIDs.ResolutionIDs) > int(a.consensusParams.Votes.MaxVotesPerTx) {
					return fmt.Errorf("number of resolution votes [%d] in a voteid transaction exceeds the limit %d", len(voteIDs.ResolutionIDs), a.consensusParams.Votes.MaxVotesPerTx)
				}
			}

			// This block proposal may include transactions that did not pass
			// through our mempool, so we have to verify all signatures.
			if !a.TxSigVerified(txO.hash) {
				if err = ident.VerifyTransaction(tx); err != nil {
					return fmt.Errorf("transaction signature verification failed: %w", err)
				}
				// We won't bother to insert this hash into the map since it is
				// very likely that this transaction is about to be executed.
			}
		}
	}
	return nil
}

// ProcessProposal should validate the received blocks and reject the block if:
// 1. transactions are not ordered by nonces
// 2. nonce is less than the last committed nonce for the account
// 3. duplicates or gaps in the nonces
// 4. transaction size is greater than the max_tx_bytes
// else accept the proposed block.
func (a *AbciApp) ProcessProposal(ctx context.Context, req *abciTypes.RequestProcessProposal) (*abciTypes.ResponseProcessProposal, error) {
	logger := a.log.With(zap.String("stage", "ABCI ProcessProposal"),
		log.Int("height", req.Height), log.Int("txs", len(req.Txs)))

	if a.forks.IsHalt(uint64(req.Height)) || a.chainContext.NetworkParameters.MigrationStatus == types.MigrationCompleted {
		if len(req.Txs) != 0 { // This network is done.  No more transactions.
			return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_REJECT}, nil
		}
		// Empty block == OK.
		return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_ACCEPT}, nil
	}

	proposerPubKey, ok := a.validatorAddressToPubKey[proposerAddrToString(req.ProposerAddress)]
	if !ok {
		// there is an edge case where cometbft will allow a node to PrepareProposal
		// even if it is not a validator, if it was a validator in the most recent block.
		// a well behaved validator will submit an empty block here, which we will accept.
		// if not an empty block, we will reject it.

		if len(req.Txs) == 0 {
			a.log.Info("proposer is not a validator and submitted an empty block, accepting proposal")
			return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_ACCEPT}, nil
		}

		a.log.Warn("proposer is not a validator and submitted a non-empty block, rejecting proposal", zap.String("proposer", proposerAddrToString(req.ProposerAddress)))
		return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_REJECT}, nil
	}

	if err := a.validateProposalTransactions(ctx, req.Txs, proposerPubKey); err != nil {
		logger.Warn("rejecting block proposal", zap.Error(err))
		return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_REJECT}, nil
	}

	// TODO: Verify the Tx and Block sizes based on the genesis configuration
	return &abciTypes.ResponseProcessProposal{Status: abciTypes.ResponseProcessProposal_ACCEPT}, nil
}

func (a *AbciApp) Query(ctx context.Context, req *abciTypes.RequestQuery) (*abciTypes.ResponseQuery, error) {
	a.log.Debug("ABCI Query", zap.String("path", req.Path), zap.String("data", string(req.Data)))
	switch {
	case req.Path == statesync.ABCISnapshotQueryPath:
		if a.snapshotter == nil {
			return nil, fmt.Errorf("this node is not configured to serve snapshots")
		}

		var snapshot *statesync.Snapshot
		height := string(req.Data)
		exists := false

		curSnapshots := a.snapshotter.ListSnapshots()
		for _, s := range curSnapshots {
			if height == fmt.Sprintf("%d", s.Height) {
				exists = true
				snapshot = s
				break
			}
		}

		if !exists {
			return nil, fmt.Errorf("snapshot not found for height %s", height)
		}

		bts, err := snapshot.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return &abciTypes.ResponseQuery{Value: bts}, nil
	case req.Path == statesync.ABCILatestSnapshotHeightPath:
		if a.snapshotter == nil {
			return nil, fmt.Errorf("this node is not configured to serve snapshots")
		}
		snaps := a.snapshotter.ListSnapshots()
		if len(snaps) == 0 {
			return nil, fmt.Errorf("no snapshots available")
		}
		latest := snaps[len(snaps)-1]
		for _, snap := range snaps {
			if snap.Height > latest.Height {
				latest = snap
			}
		}

		bts, err := latest.MarshalBinary()
		if err != nil {
			return nil, err
		}

		return &abciTypes.ResponseQuery{Value: bts}, nil
	case strings.HasPrefix(req.Path, ABCIPeerFilterPath):
		// When CometBFT connects to a peer, it sends two queries to the ABCI application
		// using the following paths, with no additional data:
		//   - `/p2p/filter/addr/<IP:PORT>`
		// 	 - `p2p/filter/id/<ID>` where ID is the peer's node ID
		// If either of these queries return a non-zero ABCI code, CometBFT will refuse to connect to the peer.
		// We manage allowed list based on the peer's node ID rather than the IP addresses, so we only need to
		// handle the `id` query and return OK for all `addr` queries.

		paths := strings.Split(req.Path[ABCIPeerFilterPathLen:], "/")
		if len(paths) != 2 {
			return nil, fmt.Errorf("invalid path: %s", req.Path)
		}

		switch paths[0] {
		case "id":
			if a.p2p.IsPeerWhitelisted(paths[1]) {
				a.log.Info("Connection attempt accepted, peer is allowed to connect", zap.String("peerID", paths[1]))
				return &abciTypes.ResponseQuery{Code: abciTypes.CodeTypeOK}, nil
			}
			// ID is not in the allowed list of peers, so reject the connection
			a.log.Warn("Connection attempt rejected, peer is not allowed to connect", zap.String("peerID", paths[1]))
			return nil, fmt.Errorf("node rejected connection attempt from peer %s", paths[1])
		case "addr":
			return &abciTypes.ResponseQuery{Code: abciTypes.CodeTypeOK}, nil
		default:
			return nil, fmt.Errorf("invalid path: %s", req.Path)
		}
	default:
		// If the query path is not recognized, return an error.
		return nil, fmt.Errorf("unknown query path: %s", req.Path)
	}
}

type EventBroadcaster func(ctx context.Context, db sql.DB, block *common.BlockContext) error

func (a *AbciApp) SetEventBroadcaster(fn EventBroadcaster) {
	a.broadcastFn = fn
}

// TxSigVerified indicates if ABCI has verified this unconfirmed transaction's
// signature. This also returns false if the transaction is not in mempool. This
// logic is not broadly applicable, but since the tx hash is computed over the
// entire serialized transaction including the signature, the same hash implies
// the same signature.
func (a *AbciApp) TxSigVerified(txHash chainHash) bool {
	a.verifiedTxnsMtx.Lock()
	defer a.verifiedTxnsMtx.Unlock()
	_, ok := a.verifiedTxns[txHash]
	return ok
}

// TxInMempool wraps TxSigVerified for callers that require a slice to check if
// a transaction is (still) in mempool.
func (a *AbciApp) TxInMempool(txHash []byte) bool {
	if len(txHash) != 32 {
		return false
	}
	hash := [32]byte(txHash) // requires go 1.20
	a.verifiedTxnsMtx.Lock()
	defer a.verifiedTxnsMtx.Unlock()
	_, ok := a.verifiedTxns[hash]
	return ok
}

// SetReplayStatusChecker sets the function to check if the node is in replay mode.
// This has to be set here because it is a CometBFT node function. Since ABCI is
// a dependency to CometBFT, this is a circular dependency, so we have to set it
// here.
func (a *AbciApp) SetReplayStatusChecker(fn func() bool) {
	a.replayingBlocks = fn
}

// Close is used to end any active database transaction that may exist if the
// application tries to shut down before closing the transaction with a call to
// Commit. Neglecting to rollback such a transaction may prevent the DB
// connection from being closed and released to the connection pool.
func (a *AbciApp) Close() error {
	if a.genesisTx != nil {
		err := a.genesisTx.Rollback(context.Background())
		if err != nil {
			return fmt.Errorf("failed to rollback genesis transaction: %w", err)
		}
	}
	if a.consensusTx != nil {
		err := a.consensusTx.Rollback(context.Background())
		if err != nil {
			return fmt.Errorf("failed to rollback consensus transaction: %w", err)
		}
	}
	return nil
}

// Price estimates the price for a transaction.
// Consumers who do not have information about the current chain parameters /
// who wanmt a guarantee that they have the most up-to-date parameters without
// reading from the DB can use this method.
func (a *AbciApp) Price(ctx context.Context, db sql.DB, tx *transactions.Transaction) (*big.Int, error) {
	return a.txApp.Price(ctx, db, tx, a.chainContext)
}

func (a *AbciApp) GetMigrationMetadata(ctx context.Context) (*types.MigrationMetadata, error) {
	a.chainContextMtx.RLock()
	defer a.chainContextMtx.RUnlock()

	status := a.chainContext.NetworkParameters.MigrationStatus
	return a.migrator.GetMigrationMetadata(ctx, status)
}

// ChangesetProcessor is a PubSub that listens for changesets and broadcasts them to the receivers.
// Subscribers can be added and removed to listen for changesets.
// Statistics receiver might listen for changesets to update the statistics every block.
// Whereas Network migrations listen for the changesets only during the migration. (that's when you register)
// ABCI --> CS Processor ---> [Subscribers]
// Once all the changesets are processed, all the channels are closed [every block]
// The channels are reset for the next block.
type changesetProcessor struct {
	// channel to receive changesets
	// closed by the pgRepl layer after all the block changes have been pushed to the processor
	csChan chan any

	// subscribers to the changeset processor are the receivers of the changesets
	// Examples: Statistics receiver, Network migration receiver
	subscribers map[string]chan<- any
}

func newChangesetProcessor() *changesetProcessor {
	return &changesetProcessor{
		csChan:      make(chan any, 1), // buffered channel to avoid blocking
		subscribers: make(map[string]chan<- any),
	}
}

// Subscribe adds a subscriber to the changeset processor's subscribers list.
// The receiver can subscribe to the changeset processor using a unique id.
func (c *changesetProcessor) Subscribe(ctx context.Context, id string) (<-chan any, error) {
	_, ok := c.subscribers[id]
	if ok {
		return nil, fmt.Errorf("subscriber with id %s already exists", id)
	}

	ch := make(chan any, 1) // buffered channel to avoid blocking
	c.subscribers[id] = ch
	return ch, nil
}

// Unsubscribe removes the subscriber from the changeset processor.
func (c *changesetProcessor) Unsubscribe(ctx context.Context, id string) error {
	if ch, ok := c.subscribers[id]; ok {
		// close the channel to signal the subscriber to stop listening
		close(ch)
		delete(c.subscribers, id)
		return nil
	}

	return fmt.Errorf("subscriber with id %s does not exist", id)
}

// Broadcast sends changesets to all the subscribers through their channels.
func (c *changesetProcessor) BroadcastChangesets(ctx context.Context) {
	defer c.Close() // All the block changesets have been processed, signal subscribers to stop listening.

	// Listen on the csChan for changesets and broadcast them to all subscribers.
	for cs := range c.csChan {
		for _, ch := range c.subscribers {
			select {
			case ch <- cs:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *changesetProcessor) Close() {
	// c.CsChan is closed by the repl layer (sender closes the channel)
	for _, ch := range c.subscribers {
		close(ch)
	}
}
