package blockprocessor

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"sort"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
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

	mtx sync.Mutex // mutex to protect the consensus params
	// consensus params
	appHash  ktypes.Hash
	height   int64
	valSet   map[string]*ktypes.Validator
	chainCtx *common.ChainContext

	// consensus TX
	consensusTx sql.PreparedTx

	// interfaces
	db          DB
	txapp       TxApp
	accounts    Accounts
	validators  ValidatorModule
	snapshotter SnapshotModule
	log         log.Logger
}

func NewBlockProcessor(ctx context.Context, db DB, txapp TxApp, accounts Accounts, vs ValidatorModule, sp SnapshotModule, genesisCfg *config.GenesisConfig, logger log.Logger) (*BlockProcessor, error) {
	// get network parameters from the chain context
	bp := &BlockProcessor{
		db:          db,
		txapp:       txapp,
		accounts:    accounts,
		validators:  vs,
		snapshotter: sp,
		log:         logger,

		valSet: make(map[string]*ktypes.Validator),
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

	height, appHash, _, err := meta.GetChainState(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain state: %w", err)
	}

	if height == -1 { // fresh initialization, initialize the chain state with the genesis params
		if err := meta.SetChainState(ctx, tx, genesisCfg.InitialHeight, []byte{}, false); err != nil {
			return nil, fmt.Errorf("failed to set chain state: %w", err)
		}

		bp.height = genesisCfg.InitialHeight
		bp.appHash = ktypes.Hash{}
	} else {
		bp.height = height
		copy(bp.appHash[:], appHash)
	}

	networkParams, err := meta.LoadParams(ctx, tx)
	if errors.Is(err, meta.ErrParamsNotFound) {
		networkParams = &common.NetworkParameters{
			MaxBlockSize:     genesisCfg.MaxBlockSize,
			JoinExpiry:       genesisCfg.JoinExpiry,
			VoteExpiry:       genesisCfg.VoteExpiry,
			DisabledGasCosts: genesisCfg.DisabledGasCosts,
			// MigrationStatus : genesisCfg.MigrationStatus,
			MaxVotesPerTx: genesisCfg.MaxVotesPerTx,
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
		// MigrationParams:   genesisCfg.MigrationParams,
	}

	// TODO: load the validator set
	validators := vs.GetValidators()
	for _, v := range validators {
		bp.valSet[hex.EncodeToString(v.PubKey)] = v
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit the outer tx: %w", err)
	}

	return bp, nil
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
	bp.height = height
	bp.appHash = appHash
	// TODO: how about validatorset and consensus params? rethink rollback

	return nil
}

// InitChain initializes the node with the genesis state. This included initializing the
// votestore with the genesis validators, accounts with the genesis allocations and the
// chain meta store with the genesis network parameters.
// This is called only once when the node is bootstrapping for the first time.
func (bp *BlockProcessor) InitChain(ctx context.Context, req *ktypes.InitChainRequest) error {
	bp.mtx.Lock()
	defer bp.mtx.Unlock()

	genesisTx, err := bp.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin the genesis transaction: %w", err)
	}
	defer genesisTx.Rollback(ctx)

	// genesis validators
	for _, v := range req.Validators {
		bp.valSet[hex.EncodeToString(v.PubKey)] = v
	}

	startParams := *bp.chainCtx.NetworkParameters

	if err := bp.txapp.GenesisInit(ctx, genesisTx, req.Validators, nil, req.InitialHeight, bp.chainCtx); err != nil {
		return err
	}

	if err := meta.SetChainState(ctx, genesisTx, req.InitialHeight, req.GenesisHash[:], false); err != nil {
		return fmt.Errorf("error storing the genesis state: %w", err)
	}

	if err := meta.StoreDiff(ctx, genesisTx, &startParams, bp.chainCtx.NetworkParameters); err != nil {
		return fmt.Errorf("error storing the genesis consensus params: %w", err)
	}

	// TODO: Genesis hash and what are the mechanics for producing the first block (genesis block)?
	bp.txapp.Commit()

	bp.log.Infof("Initialized chain: height %d, appHash: %s", req.InitialHeight, req.GenesisHash.String())

	return genesisTx.Commit(ctx)
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
		return nil, fmt.Errorf("failed to begin the consensus transaction: %w", err)
	}

	blockCtx := &common.BlockContext{
		Height:       req.Height,
		Timestamp:    req.Block.Header.Timestamp.Unix(),
		ChainContext: bp.chainCtx,
		Proposer:     req.Proposer,
	}

	txResults := make([]ktypes.TxResult, len(req.Block.Txns))
	for i, tx := range req.Block.Txns {
		decodedTx := &ktypes.Transaction{}
		if err := decodedTx.UnmarshalBinary(tx); err != nil {
			// bp.log.Error("Failed to unmarshal the block tx", "err", err)
			return nil, fmt.Errorf("failed to unmarshal the block tx: %w", err)
		}
		txHash := types.HashBytes(tx)

		auth := auth.GetAuthenticator(decodedTx.Signature.Type)

		identifier, err := auth.Identifier(decodedTx.Sender)
		if err != nil {
			// bp.log.Error("Failed to get identifier for the block tx", "err", err)
			return nil, fmt.Errorf("failed to get identifier for the block tx: %w", err)
		}

		txCtx := &common.TxContext{
			Ctx:           ctx,
			TxID:          hex.EncodeToString(txHash[:]),
			Signer:        decodedTx.Sender,
			Authenticator: decodedTx.Signature.Type,
			Caller:        identifier,
			BlockContext:  blockCtx,
		}

		select {
		case <-ctx.Done(): // TODO: is this the best way to abort the block execution?
			bp.log.Info("Block execution cancelled", "height", req.Height)
			return nil, nil // TODO: or error? or trigger resetState?
		default:
			res := bp.txapp.Execute(txCtx, bp.consensusTx, decodedTx)
			txResult := ktypes.TxResult{
				Code: uint32(res.ResponseCode),
				Gas:  res.Spend,
			}
			if res.Error != nil {
				if sql.IsFatalDBError(res.Error) {
					return nil, fmt.Errorf("fatal db error during block execution: %w", res.Error)
				}

				txResult.Log = res.Error.Error()
				bp.log.Debug("Failed to execute transaction", "tx", txHash, "err", res.Error)
			} else {
				txResult.Log = "success"
			}

			txResults[i] = txResult
		}
	}

	_, err = bp.txapp.Finalize(ctx, bp.consensusTx, blockCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize the block execution: %w", err)
	}

	if err := meta.SetChainState(ctx, bp.consensusTx, req.Height, bp.appHash[:], true); err != nil {
		return nil, fmt.Errorf("failed to set the chain state: %w", err)
	}

	// Create a new changeset processor
	csp := newChangesetProcessor()
	// "migrator" module subscribes to the changeset processor to store changesets during the migration
	csErrChan := make(chan error, 1)
	defer close(csErrChan)
	// TODO: Subscribe to the changesets
	go csp.BroadcastChangesets(ctx)

	appHash, err := bp.consensusTx.Precommit(ctx, csp.csChan)
	if err != nil {
		return nil, fmt.Errorf("failed to precommit the changeset: %w", err)
	}

	valUpdates := bp.validators.ValidatorUpdates()
	valUpdatesList := make([]*ktypes.Validator, 0) // TODO: maybe change the validatorUpdates API to return a list instead of map
	valUpdatesHash := validatorUpdatesHash(valUpdates)
	for k, v := range valUpdates {
		if v.Power == 0 {
			delete(bp.valSet, k)
		} else {
			bp.valSet[k] = &ktypes.Validator{
				PubKey: v.PubKey,
				Power:  v.Power,
			}
		}
		valUpdatesList = append(valUpdatesList, v)
	}

	accountsHash := bp.accountsHash()
	txResultsHash := txResultsHash(txResults)

	nextHash := bp.nextAppHash(bp.appHash, types.Hash(appHash), valUpdatesHash, accountsHash, txResultsHash)

	bp.height = req.Height
	bp.appHash = nextHash

	bp.log.Info("Executed Block", "height", req.Height, "blkHash", req.BlockID, "appHash", nextHash)

	return &ktypes.BlockExecResult{
		TxResults:        txResults,
		AppHash:          nextHash,
		ValidatorUpdates: valUpdatesList,
	}, nil

}

// Commit method commits the block to the blockstore and postgres database.
// It also updates the txIndexer and mempool with the transactions in the block.
func (bp *BlockProcessor) Commit(ctx context.Context, height int64, appHash ktypes.Hash, syncing bool) error {
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

	if err := meta.SetChainState(ctxS, tx, height, appHash[:], false); err != nil {
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
	if err := bp.snapshotDB(ctx, height, syncing); err != nil {
		bp.log.Warn("Failed to create snapshot of the database", "err", err)
	}

	bp.log.Info("Committed Block", "height", height, "appHash", appHash.String())
	return nil
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
		return bytes.Compare(a.Identifier, b.Identifier)
	})
	hasher := ktypes.NewHasher()
	for _, acc := range accounts {
		hasher.Write(acc.Identifier)
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

// We probably don't want to do this long term
func (bp *BlockProcessor) Price(ctx context.Context, dbTx sql.DB, tx *ktypes.Transaction) (*big.Int, error) {
	return bp.txapp.Price(ctx, dbTx, tx, bp.chainCtx)
}

func (bp *BlockProcessor) AccountInfo(ctx context.Context, db sql.DB, identifier []byte, pending bool) (balance *big.Int, nonce int64, err error) {
	return bp.txapp.AccountInfo(ctx, db, identifier, pending)
}
