package blockprocessor

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	authExt "github.com/kwilteam/kwil-db/extensions/auth"
	"github.com/kwilteam/kwil-db/node/ident"
	"github.com/kwilteam/kwil-db/node/meta"
	"github.com/kwilteam/kwil-db/node/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// This package will be equivalent to the ABCI application in Tendermint.
// This is responsible for processing blocks, managing consensus state, and
// handling transactions and mempool state.
// Once Consensus Engine decides on the block to be processed, it will be sent to
// Block Processor for its execution.
// Modules like Consensus Engine, RPC services rely on this package.

// shouldn't be generally concerned with the role. maybe role is needed for broadcasting eventstore events to the leader.
type BlockProcessor struct {
	// config
	genesisParams *config.GenesisConfig
	signer        auth.Signer

	mtx sync.RWMutex // mutex to protect the consensus params
	// consensus params
	appHash  ktypes.Hash
	height   atomic.Int64
	chainCtx *common.ChainContext

	status   *blockExecStatus
	statusMu sync.RWMutex // very granular mutex to protect access to the block execution status

	// consensus TX
	consensusTx sql.PreparedTx

	// interfaces
	db          DB
	txapp       TxApp
	accounts    Accounts
	validators  ValidatorModule
	snapshotter SnapshotModule
	events      EventStore
	migrator    MigratorModule
	log         log.Logger

	broadcastTxFn BroadcastTxFn

	// Subscribers for the validator updates
	subChans []chan []*ktypes.Validator
	subMtx   sync.RWMutex
}

type BroadcastTxFn func(ctx context.Context, tx *ktypes.Transaction, sync uint8) (*ktypes.ResultBroadcastTx, error)

func NewBlockProcessor(ctx context.Context, db DB, txapp TxApp, accounts Accounts, vs ValidatorModule, sp SnapshotModule, es EventStore, migrator MigratorModule, bs BlockStore, genesisCfg *config.GenesisConfig, signer auth.Signer, logger log.Logger) (*BlockProcessor, error) {
	// get network parameters from the chain context
	bp := &BlockProcessor{
		db:          db,
		txapp:       txapp,
		accounts:    accounts,
		validators:  vs,
		snapshotter: sp,
		signer:      signer,
		events:      es,
		migrator:    migrator,
		log:         logger,
	}

	if genesisCfg == nil { // TODO: remove this
		genesisCfg = config.DefaultGenesisConfig()
	}

	bp.genesisParams = genesisCfg

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin outer tx: %w", err)
	}
	defer tx.Rollback(ctx)

	height, appHash, dirty, err := meta.GetChainState(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain state: %w", err)
	}
	if dirty {
		// app state is in a partially committed state, recover the chain state.
		_, _, ci, err := bs.GetByHeight(height)
		if err != nil || ci == nil {
			return nil, err
		}

		if err := meta.SetChainState(ctx, tx, height, ci.AppHash[:], false); err != nil {
			return nil, err
		}

		copy(appHash, ci.AppHash[:])

		// also update the last changeset height in the migrator
		if err := bp.migrator.PersistLastChangesetHeight(ctx, tx, height); err != nil {
			return nil, err
		}
	}

	bp.height.Store(height)
	copy(bp.appHash[:], appHash)

	var migrationParams *common.MigrationContext
	startHeight := genesisCfg.Migration.StartHeight
	endHeight := genesisCfg.Migration.EndHeight

	if startHeight != 0 && endHeight != 0 {
		migrationParams = &common.MigrationContext{
			StartHeight: startHeight,
			EndHeight:   endHeight,
		}
	}

	networkParams, err := meta.LoadParams(ctx, tx)
	if errors.Is(err, meta.ErrParamsNotFound) {
		status := ktypes.NoActiveMigration
		if migrationParams != nil {
			status = ktypes.GenesisMigration
		}

		networkParams = &common.NetworkParameters{
			MaxBlockSize:     genesisCfg.MaxBlockSize,
			JoinExpiry:       genesisCfg.JoinExpiry,
			VoteExpiry:       genesisCfg.VoteExpiry,
			DisabledGasCosts: genesisCfg.DisabledGasCosts,
			MigrationStatus:  status,
			MaxVotesPerTx:    genesisCfg.MaxVotesPerTx,
		}

		if err := meta.StoreParams(ctx, tx, networkParams); err != nil {
			return nil, fmt.Errorf("failed to store the network parameters: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to load the network parameters: %w", err)
	}

	bp.chainCtx = &common.ChainContext{
		ChainID:           genesisCfg.ChainID,
		NetworkParameters: networkParams,
		MigrationParams:   migrationParams,
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit the outer tx: %w", err)
	}

	return bp, nil
}

func (bp *BlockProcessor) SetBroadcastTxFn(fn BroadcastTxFn) {
	bp.broadcastTxFn = fn
}

func (bp *BlockProcessor) Close() error {
	bp.mtx.Lock()
	defer bp.mtx.Unlock()

	if bp.consensusTx != nil {
		bp.log.Info("Rolling back the consensus transaction")
		if err := bp.consensusTx.Rollback(context.Background()); err != nil {
			return fmt.Errorf("failed to rollback the consensus transaction: %w", err)
		}
	}

	return nil
}

func (bp *BlockProcessor) Rollback(ctx context.Context, height int64, appHash ktypes.Hash) error {
	bp.mtx.Lock()
	defer bp.mtx.Unlock()

	if bp.consensusTx != nil {
		bp.log.Info("Rolling back the consensus transaction")
		if err := bp.consensusTx.Rollback(context.Background()); err != nil {
			return fmt.Errorf("failed to rollback the consensus transaction: %w", err)
		}
	}

	// set the block proposer back to it's previous state
	bp.height.Store(height)
	bp.appHash = appHash

	readTx, err := bp.db.BeginReadTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read transaction: %w", err)
	}
	defer readTx.Rollback(ctx)

	networkParams, err := meta.LoadParams(ctx, readTx)
	if err != nil {
		return fmt.Errorf("failed to load the network parameters: %w", err)
	}

	bp.chainCtx.NetworkParameters = networkParams

	// Rollback internal state updates to the validators, accounts and mempool.
	bp.txapp.Rollback()

	return nil
}

func (bp *BlockProcessor) CheckTx(ctx context.Context, tx *ktypes.Transaction, recheck bool) error {
	txHash := tx.Hash()

	// If the network is halted for migration, we reject all transactions.
	if bp.chainCtx.NetworkParameters.MigrationStatus == ktypes.MigrationCompleted {
		return fmt.Errorf("network is halted for migration")
	}

	bp.log.Info("Check transaction", "Recheck", recheck, "Hash", txHash, "Sender", hex.EncodeToString(tx.Sender),
		"PayloadType", tx.Body.PayloadType.String(), "Nonce", tx.Body.Nonce, "TxFee", tx.Body.Fee.String())

	if !recheck {
		// Verify the correct chain ID is set, if it is set.
		if protected := tx.Body.ChainID != ""; protected && tx.Body.ChainID != bp.genesisParams.ChainID {
			bp.log.Info("Wrong chain ID", "txChainID", tx.Body.ChainID)
			return fmt.Errorf("wrong chain ID: %s", tx.Body.ChainID)
		}

		// Ensure that the transaction is valid in terms of the signature and the payload type
		if err := ident.VerifyTransaction(tx); err != nil {
			bp.log.Debug("Failed to verify the transaction", "err", err)
			return fmt.Errorf("failed to verify the transaction: %w", err)
		}
	}

	readTx, err := bp.db.BeginReadTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read transaction: %w", err)
	}
	defer readTx.Rollback(ctx)

	auth, err := authExt.GetAuthenticator(tx.Signature.Type)
	if err != nil {
		return fmt.Errorf("failed to get authenticator: %w", err)
	}

	ident, err := auth.Identifier(tx.Sender)
	if err != nil {
		return fmt.Errorf("failed to get identifier: %w", err)
	}

	err = bp.txapp.ApplyMempool(&common.TxContext{
		Ctx: ctx,
		BlockContext: &common.BlockContext{
			ChainContext: bp.chainCtx,
			Height:       bp.height.Load() + 1,
			Proposer:     bp.genesisParams.Leader, // always the leader?
		},
		TxID:          hex.EncodeToString(txHash[:]),
		Signer:        tx.Sender,
		Caller:        ident,
		Authenticator: tx.Signature.Type,
	}, readTx, tx)
	if err != nil {
		// do appropriate logging
		bp.log.Info("Failed to apply the transaction to the mempool", "tx", hex.EncodeToString(txHash[:]), "err", err)
		return err
	}

	return nil
}

// InitChain initializes the node with the genesis state. This included initializing the
// votestore with the genesis validators, accounts with the genesis allocations and the
// chain meta store with the genesis network parameters.
// This is called only once when the node is bootstrapping for the first time.
func (bp *BlockProcessor) InitChain(ctx context.Context) (int64, []byte, error) {
	bp.mtx.Lock()
	defer bp.mtx.Unlock()

	genesisTx, err := bp.db.BeginTx(ctx)
	if err != nil {
		return -1, nil, fmt.Errorf("failed to begin the genesis transaction: %w", err)
	}
	defer genesisTx.Rollback(ctx)

	genCfg := bp.genesisParams

	if err := bp.txapp.GenesisInit(ctx, genesisTx, genCfg.Validators, nil, genCfg.InitialHeight, genCfg.DBOwner, bp.chainCtx); err != nil {
		return -1, nil, err
	}

	if err := meta.SetChainState(ctx, genesisTx, genCfg.InitialHeight, genCfg.StateHash, false); err != nil {
		return -1, nil, fmt.Errorf("error storing the genesis state: %w", err)
	}

	if err := bp.txapp.Commit(); err != nil {
		return -1, nil, fmt.Errorf("txapp commit failed: %w", err)
	}

	if err := genesisTx.Commit(ctx); err != nil {
		return -1, nil, fmt.Errorf("genesis transaction commit failed: %w", err)
	}

	bp.announceValidators()

	bp.height.Store(genCfg.InitialHeight)
	copy(bp.appHash[:], genCfg.StateHash)

	bp.log.Infof("Initialized chain: height %d, appHash: %s", genCfg.InitialHeight, hex.EncodeToString(genCfg.StateHash))

	return genCfg.InitialHeight, genCfg.StateHash, nil
}

func (bp *BlockProcessor) ExecuteBlock(ctx context.Context, req *ktypes.BlockExecRequest) (blkResult *ktypes.BlockExecResult, err error) {
	bp.mtx.Lock()
	defer bp.mtx.Unlock()

	// Begin the block execution session
	if err = bp.txapp.Begin(ctx, req.Height); err != nil {
		return nil, fmt.Errorf("failed to begin the block execution: %w", err)
	}

	bp.consensusTx, err = bp.db.BeginPreparedTx(ctx)
	if err != nil {
		bp.consensusTx = nil // safety measure
		return nil, fmt.Errorf("failed to begin the consensus transaction: %w", err)
	}

	// we copy the Kwil consensus params to ensure we persist any changes
	// made during the block execution
	networkParams := &common.NetworkParameters{
		MaxBlockSize:     bp.chainCtx.NetworkParameters.MaxBlockSize,
		JoinExpiry:       bp.chainCtx.NetworkParameters.JoinExpiry,
		VoteExpiry:       bp.chainCtx.NetworkParameters.VoteExpiry,
		DisabledGasCosts: bp.chainCtx.NetworkParameters.DisabledGasCosts,
		MaxVotesPerTx:    bp.chainCtx.NetworkParameters.MaxVotesPerTx,
		MigrationStatus:  bp.chainCtx.NetworkParameters.MigrationStatus,
	}
	oldNetworkParams := *networkParams

	blockCtx := &common.BlockContext{
		Height:       req.Height,
		Timestamp:    req.Block.Header.Timestamp.Unix(),
		ChainContext: bp.chainCtx,
		Proposer:     req.Proposer,
	}

	inMigration := blockCtx.ChainContext.NetworkParameters.MigrationStatus == ktypes.MigrationInProgress
	haltNetwork := blockCtx.ChainContext.NetworkParameters.MigrationStatus == ktypes.MigrationCompleted

	txResults := make([]ktypes.TxResult, len(req.Block.Txns))

	txHashes := bp.initBlockExecutionStatus(req.Block)

	for i, tx := range req.Block.Txns {
		auth := auth.GetAuthenticator(tx.Signature.Type)
		if auth == nil {
			return nil, fmt.Errorf("unsupported signature type: %v", tx.Signature.Type)
		}

		identifier, err := auth.Identifier(tx.Sender)
		if err != nil {
			return nil, fmt.Errorf("failed to get identifier for the block tx: %w", err)
		}

		txHash := txHashes[i]

		txCtx := &common.TxContext{
			Ctx:           ctx,
			TxID:          txHash.String(),
			Signer:        tx.Sender,
			Authenticator: tx.Signature.Type,
			Caller:        identifier,
			BlockContext:  blockCtx,
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err() // notify the caller about the context cancellation or deadline exceeded error
		default:
			res := bp.txapp.Execute(txCtx, bp.consensusTx, tx)
			txResult := ktypes.TxResult{
				Code: uint32(res.ResponseCode),
				Gas:  res.Spend,
			}

			// bookkeeping for the block execution status
			bp.updateBlockExecutionStatus(txHash)

			if res.Error != nil {
				if sql.IsFatalDBError(res.Error) {
					return nil, fmt.Errorf("fatal db error during block execution: %w", res.Error)
				}

				txResult.Log = res.Error.Error()
				bp.log.Info("Failed to execute transaction", "tx", txHash, "err", res.Error)
			} else {
				txResult.Log = "success"
			}

			txResults[i] = txResult
		}
	}

	// record the end time of the block execution
	bp.recordBlockExecEndTime()

	// Broadcast any voteID events that have not been broadcasted yet
	if bp.broadcastTxFn != nil {
		if err = bp.BroadcastVoteIDTx(ctx, bp.consensusTx); err != nil {
			return nil, fmt.Errorf("failed to broadcast the voteID transactions: %w", err)
		}
	}

	_, err = bp.txapp.Finalize(ctx, bp.consensusTx, blockCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize the block execution: %w", err)
	}

	// migrator can be updated here within notify height
	err = bp.migrator.NotifyHeight(ctx, blockCtx, bp.db)
	if err != nil {
		return nil, fmt.Errorf("failed to notify the migrator about the block height: %w", err)
	}

	networkParams.MigrationStatus = bp.chainCtx.NetworkParameters.MigrationStatus

	if err := meta.SetChainState(ctx, bp.consensusTx, req.Height, bp.appHash[:], true); err != nil {
		return nil, fmt.Errorf("failed to set the chain state: %w", err)
	}

	if err := meta.StoreDiff(ctx, bp.consensusTx, &oldNetworkParams, bp.chainCtx.NetworkParameters); err != nil {
		return nil, fmt.Errorf("failed to store the network parameters: %w", err)
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
			csErrChan <- bp.migrator.StoreChangesets(req.Height, csChanMigrator)
		}()
	}

	go csp.BroadcastChangesets(ctx)

	appHash, err := bp.consensusTx.Precommit(ctx, csp.csChan)
	if err != nil {
		return nil, fmt.Errorf("failed to precommit the changeset: %w", err)
	}

	valUpdates := bp.validators.ValidatorUpdates()
	valUpdatesList := make([]*ktypes.Validator, 0) // TODO: maybe change the validatorUpdates API to return a list instead of map
	valUpdatesHash := validatorUpdatesHash(valUpdates)
	for _, v := range valUpdates {
		valUpdatesList = append(valUpdatesList, v)
	}

	accountsHash := bp.accountsHash()
	txResultsHash := txResultsHash(txResults)

	nextHash := bp.nextAppHash(bp.appHash, types.Hash(appHash), valUpdatesHash, accountsHash, txResultsHash)

	if inMigration && !haltNetwork {
		// wait for the migrator to finish storing the changesets
		err = <-csErrChan
		if err != nil {
			return nil, fmt.Errorf("failed to store changesets during migration: %w", err)
		}
	}

	bp.log.Info("Executed Block", "height", req.Height, "blkHash", req.BlockID, "appHash", nextHash)

	return &ktypes.BlockExecResult{
		TxResults:        txResults,
		AppHash:          nextHash,
		ValidatorUpdates: valUpdatesList,
	}, nil

}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (bp *BlockProcessor) Commit(ctx context.Context, req *ktypes.CommitRequest) error {
	bp.mtx.Lock()
	defer bp.mtx.Unlock()

	// Commit the Postgres Consensus transaction
	if err := bp.consensusTx.Commit(ctx); err != nil {
		return err
	}
	bp.consensusTx = nil

	// Update the chain meta store with the new height and the dirty
	// we need to re-open a new transaction just to write the apphash
	// TODO: it would be great to have a way to commit the apphash without
	// opening a new transaction. This could leave us in a state where data is
	// committed but the apphash is not, which would essentially nuke the chain.
	ctxS := context.Background()
	tx, err := bp.db.BeginTx(ctxS) // badly timed shutdown MUST NOT cancel now, we need consistency with consensus tx commit
	if err != nil {
		return err
	}

	if err := meta.SetChainState(ctxS, tx, req.Height, req.AppHash[:], false); err != nil {
		err2 := tx.Rollback(ctxS)
		if err2 != nil {
			bp.log.Error("Failed to rollback the transaction", "err", err2)
			return err2
		}
		return err
	}

	if err := bp.migrator.PersistLastChangesetHeight(ctxS, tx, req.Height); err != nil {
		err2 := tx.Rollback(ctxS)
		if err2 != nil {
			bp.log.Error("Failed to rollback the transaction", "err", err2)
			return err2
		}
		return err
	}

	if err := tx.Commit(ctxS); err != nil {
		return err
	}

	if err := bp.txapp.Commit(); err != nil {
		return err
	}

	// Snapshots:
	if err := bp.snapshotDB(ctx, req.Height, req.Syncing); err != nil {
		bp.log.Warn("Failed to create snapshot of the database", "err", err)
	}

	bp.clearBlockExecutionStatus() // TODO: not very sure where to clear this

	bp.height.Store(req.Height)
	copy(bp.appHash[:], req.AppHash[:])

	// Announce final validators to subscribers
	bp.announceValidators()

	bp.log.Info("Committed Block", "height", req.Height, "appHash", req.AppHash.String())
	return nil
}

// This function enforces proper nonce ordering, validates transactions, and ensures
// that consensus limits such as the maximum block size, maxVotesPerTx are met. It also adds
// validator vote transactions for events observed by the leader. This function is
// used exclusively by the leader node to prepare the proposal block.
func (bp *BlockProcessor) PrepareProposal(ctx context.Context, txs []*ktypes.Transaction) (finalTxs []*ktypes.Transaction, invalidTxs []*ktypes.Transaction, err error) {
	// unmarshal and index the transactions
	return bp.prepareBlockTransactions(ctx, txs)
}

var (
	statesyncSnapshotSchemas = []string{"kwild_voting", "kwild_internal", "kwild_chain", "kwild_accts", "kwild_migrations", "ds_*"}
	statsyncExcludedTables   = []string{"kwild_internal.sentry"}
)

func (bp *BlockProcessor) snapshotDB(ctx context.Context, height int64, syncing bool) error {
	snapshotsDue := bp.snapshotter.Enabled() &&
		(bp.snapshotter.IsSnapshotDue(uint64(height)) || len(bp.snapshotter.ListSnapshots()) == 0)
	// snapshotsDue = snapshotsDue && height > max(1, a.cfg.InitialHeight)

	if snapshotsDue && !syncing {
		// we make a snapshot tx but don't directly use it. This is because under the hood,
		// we are using the pg_dump executable to create the snapshot, and we are simply
		// giving pg_dump the snapshot ID to guarantee it has an isolated view of the database.
		snapshotTx, snapshotId, err := bp.db.BeginSnapshotTx(ctx)
		if err != nil {
			return fmt.Errorf("failed to start snapshot tx: %w", err)
		}
		defer snapshotTx.Rollback(ctx) // always rollback, since this is just for view isolation

		err = bp.snapshotter.CreateSnapshot(ctx, uint64(height), snapshotId, statesyncSnapshotSchemas, statsyncExcludedTables, nil)
		if err != nil {
			return err
		} else {
			bp.log.Info("created snapshot", "height", height, "snapshot_id", snapshotId)
			return nil
		}
	}

	return nil
}

// nextAppHash calculates the appHash that encapsulates the state changes occurred during the block execution.
// sha256(prevAppHash || changesetHash || valUpdatesHash || accountsHash || txResultsHash)
func (bp *BlockProcessor) nextAppHash(prevAppHash, changesetHash, valUpdatesHash, accountsHash, txResultsHash types.Hash) types.Hash {
	hasher := ktypes.NewHasher()

	hasher.Write(prevAppHash[:])
	hasher.Write(changesetHash[:])
	hasher.Write(valUpdatesHash[:])
	hasher.Write(accountsHash[:])
	hasher.Write(txResultsHash[:])

	bp.log.Info("AppState updates: ", "prevAppHash", prevAppHash, "changesetsHash", changesetHash, "valUpdatesHash", valUpdatesHash, "accountsHash", accountsHash, "txResultsHash", txResultsHash)
	return hasher.Sum(nil)
}

func txResultsHash(results []ktypes.TxResult) types.Hash {
	hasher := ktypes.NewHasher()
	for _, res := range results {
		binary.Write(hasher, binary.BigEndian, res.Code)
		binary.Write(hasher, binary.BigEndian, res.Gas)
	}

	return hasher.Sum(nil)
}

func (bp *BlockProcessor) accountsHash() types.Hash {
	accounts := bp.accounts.Updates()
	slices.SortFunc(accounts, func(a, b *ktypes.Account) int {
		return strings.Compare(a.Identifier, b.Identifier)
	})
	hasher := ktypes.NewHasher()
	for _, acc := range accounts {
		hasher.Write([]byte(acc.Identifier))
		binary.Write(hasher, binary.BigEndian, acc.Balance.Bytes())
		binary.Write(hasher, binary.BigEndian, acc.Nonce)
	}

	return hasher.Sum(nil)
}

func validatorUpdatesHash(updates map[string]*ktypes.Validator) types.Hash {
	// sort the updates by the validator address
	// hash the validator address and the validator struct

	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hash := ktypes.NewHasher()
	for _, k := range keys {
		// hash the validator address
		hash.Write(updates[k].PubKey)
		// hash the validator power
		binary.Write(hash, binary.BigEndian, updates[k].Power)
	}

	return hash.Sum(nil)
}

// SubscribeValidators creates and returns a new channel on which the current
// validator set will be sent for each block Commit. The receiver will miss
// updates if they are unable to receive fast enough. This should generally
// be used after catch-up is complete, and only called once by the receiving
// goroutine rather than repeatedly in a loop, for instance. The slice should
// not be modified by the receiver.
func (bp *BlockProcessor) SubscribeValidators() <-chan []*ktypes.Validator {
	// There's only supposed to be one user of this method, and they should
	// only get one channel and listen, but play it safe and use a slice.
	bp.subMtx.Lock()
	defer bp.subMtx.Unlock()

	c := make(chan []*ktypes.Validator, 1)
	bp.subChans = append(bp.subChans, c)
	return c
}

// announceValidators sends the current validator list to subscribers from
// ReceiveValidators.
func (bp *BlockProcessor) announceValidators() {
	// dev note: this method should not be blocked by receivers. Keep a default
	// case and create buffered channels.
	bp.subMtx.RLock()
	defer bp.subMtx.RUnlock()

	if len(bp.subChans) == 0 {
		return // no subscribers, skip the slice clone
	}

	vals := bp.GetValidators()

	for _, c := range bp.subChans {
		select {
		case c <- vals:
		default: // they'll get the next one... this is just supposed to be better than polling
			bp.log.Warn("Validator update channel is blocking")
		}
	}
}

func (bp *BlockProcessor) Price(ctx context.Context, dbTx sql.DB, tx *ktypes.Transaction) (*big.Int, error) {
	return bp.txapp.Price(ctx, dbTx, tx, bp.chainCtx)
}

func (bp *BlockProcessor) AccountInfo(ctx context.Context, db sql.DB, identifier string, pending bool) (balance *big.Int, nonce int64, err error) {
	return bp.txapp.AccountInfo(ctx, db, identifier, pending)
}

func (bp *BlockProcessor) GetValidators() []*ktypes.Validator {
	return bp.validators.GetValidators()
}

func (bp *BlockProcessor) ConsensusParams() *ktypes.ConsensusParams {
	bp.mtx.RLock()
	defer bp.mtx.RUnlock()

	return &ktypes.ConsensusParams{
		MaxBlockSize:     bp.chainCtx.NetworkParameters.MaxBlockSize,
		JoinExpiry:       bp.chainCtx.NetworkParameters.JoinExpiry,
		VoteExpiry:       bp.chainCtx.NetworkParameters.VoteExpiry,
		DisabledGasCosts: bp.chainCtx.NetworkParameters.DisabledGasCosts,
		MaxVotesPerTx:    bp.chainCtx.NetworkParameters.MaxVotesPerTx,
		MigrationStatus:  bp.chainCtx.NetworkParameters.MigrationStatus,
	}
}

func (bp *BlockProcessor) SetNetworkParameters(params *common.NetworkParameters) {
	bp.mtx.Lock()
	defer bp.mtx.Unlock()

	bp.chainCtx.NetworkParameters = &common.NetworkParameters{
		MaxBlockSize:     params.MaxBlockSize,
		JoinExpiry:       params.JoinExpiry,
		VoteExpiry:       params.VoteExpiry,
		DisabledGasCosts: params.DisabledGasCosts,
		MaxVotesPerTx:    params.MaxVotesPerTx,
		MigrationStatus:  params.MigrationStatus,
	}
}

func (bp *BlockProcessor) GetMigrationMetadata(ctx context.Context) (*ktypes.MigrationMetadata, error) {
	bp.mtx.RLock()
	status := bp.chainCtx.NetworkParameters.MigrationStatus
	bp.mtx.RUnlock()

	return bp.migrator.GetMigrationMetadata(ctx, status)
}
