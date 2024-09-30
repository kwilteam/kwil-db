package migrations

import (
	"context"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	cmtTypes "github.com/cometbft/cometbft/types"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
	"github.com/kwilteam/kwil-db/internal/txapp"
	"github.com/kwilteam/kwil-db/internal/voting"
)

var (
	ErrNoActiveMigration = fmt.Errorf("no active migration")

	// Schemas to include in the network migration snapshot.
	networkMigrationSchemas = []string{
		"kwild_voting",
		"kwild_accts",
		"kwild_internal",
		"ds_*",
	}

	// Tables to exclude from the network migration snapshot.
	networkMigrationExcludedTables = []string{"kwild_internal.sentry"}

	// Tables for which data should be excluded from the network migration snapshot.
	networkMigrationExcludedTableData = []string{
		"kwild_voting.voters",
	}

	metadataFileName = "metadata.json"
)

const (
	MaxChunkSize = 1 * 1000 * 1000 // around 1MB
)

// migrator is responsible for managing the migrations.
// It is responsible for tracking any in-process migrations, snapshotting
// at the appropriate height, persisting changesets for the migration for each
// block as it occurs, and making that data available via RPC for the new node.
// Similarly, if the local process is the new node, it is responsible for reading
// changesets from the external node and applying them to the local database.
// The changesets are stored from the start height of the migration to the end height (both inclusive).
type Migrator struct {
	initialized bool // set to true after the migrator is initialized

	// mu is the mutex for the migrator.
	mu sync.RWMutex

	// activeMigration is the migration plan that is approved by the network.
	// It is nil if there is no plan for a migration.
	activeMigration *activeMigration

	// snapshotter creates snapshots of the state.
	snapshotter Snapshotter

	// DB is a connection to the database.
	// It should connect to the same Postgres database as kwild,
	// but should be a different connection pool.
	DB Database

	// accounts tracks all the spends that have occurred in the block.
	accounts SpendTracker

	// lastChangeset is the height of the last changeset that was stored.
	// If no changesets have been stored, it is -1.
	lastChangeset int64

	// Logger is the logger for the migrator.
	Logger log.Logger

	// dir is the directory where the migration data is stored.
	// It is expected to be a full path.
	dir string

	// consensusParamsFn is a function that returns the consensus params for the chain.
	consensusParamsFn ConsensusParamsGetter
	// consensusParamsFnChan is a channel that is signals if the consensusParamsFn is set.
	consensusParamsFnChan chan struct{}

	genesisMigrationParams chain.MigrationParams

	migrationStatus types.MigrationStatus
}

// activeMigration is an in-process migration.
type activeMigration struct {
	// StartHeight is the height at which the migration starts.
	StartHeight int64
	// EndHeight is the height at which the migration ends.
	EndHeight int64
}

// SetupMigrator initializes the migrator instance with the necessary dependencies.
func SetupMigrator(ctx context.Context, db Database, snapshotter Snapshotter, accounts SpendTracker, dir string, migrationParams chain.MigrationParams, logger log.Logger) (*Migrator, error) {
	if migrator.initialized {
		return nil, fmt.Errorf("migrator already initialized")
	}

	// Set the migrator declared in migrations.go
	migrator.genesisMigrationParams = migrationParams
	migrator.snapshotter = snapshotter
	migrator.Logger = logger
	migrator.dir = dir
	migrator.DB = db
	migrator.accounts = accounts
	migrator.initialized = true
	migrator.consensusParamsFnChan = make(chan struct{})
	// Initialize the DB
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initializeMigrationSchema,
	}
	err := versioning.Upgrade(ctx, db, migrationsSchemaName, upgradeFns, migrationSchemaVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade migrations DB: %w", err)
	}

	// retrieve migration metadata
	tx, err := db.BeginReadTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin read tx: %w", err)
	}
	defer tx.Rollback(ctx)

	m, err := getMigrationState(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration state: %w", err)
	}
	migrator.activeMigration = m

	// Get the last changeset that was stored
	height, err := getLastStoredChangeset(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get last stored changeset: %w", err)
	}
	migrator.lastChangeset = height

	return migrator, nil
}

// NotifyHeight notifies the migrator that a new block has been committed.
// It is called at the end of the block being applied, but before the block is
// committed to the database, in between tx.PreCommit and tx.Commit.
func (m *Migrator) NotifyHeight(ctx context.Context, block *common.BlockContext, db Database) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if there is no active migration, there is nothing to do
	if m.activeMigration == nil {
		return nil
	}

	// if not in a migration, we can return early
	if block.Height < m.activeMigration.StartHeight-1 {
		return nil
	}

	if block.Height > m.activeMigration.EndHeight {
		return nil
	}

	if block.Height == m.activeMigration.StartHeight-1 {
		// set the migration in progress, so that we record the changesets starting from the next block
		block.ChainContext.NetworkParameters.MigrationStatus = types.MigrationInProgress
		return nil
	}

	/*
		I previously thought to make this run asynchronously, since PG dump can take a significant amount of time,
		however I decided againast it, because nodes are required to agree on the height of the old chain during the
		migration on the new chain. Im not sure of a way to guarantee this besdies literally enforcing that the old
		chain runs the migration synchronously as part of consensus.

		NOTE: https://github.com/kwilteam/kwil-db/pull/837#discussion_r1648036539
	*/

	// if the current block height is the height at which the migration starts, then
	// we should snapshot the current DB and begin the migration. Since NotifyHeight is called
	// during PreCommit, the state changes from the current block won't be included in the snapshot,
	// and will instead need to be recorded as the first changeset of the migration.

	if block.Height == m.activeMigration.StartHeight {
		tx, snapshotId, err := db.BeginSnapshotTx(ctx)
		if err != nil {
			return err
		}
		defer func() {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				// we can mostly ignore this error, since the original err will halt the node anyways
				m.Logger.Errorf("failed to rollback transaction: %s", err2.Error())
			}
		}()

		err = m.snapshotter.CreateSnapshot(ctx, uint64(block.Height), snapshotId, networkMigrationSchemas, networkMigrationExcludedTables, networkMigrationExcludedTableData)
		if err != nil {
			return err
		}

		// Generate a genesis file for the snapshot
		vals, err := voting.GetValidators(ctx, tx)
		if err != nil {
			return err
		}

		// Retrieve snapshot hash
		snapshots := m.snapshotter.ListSnapshots()
		if len(snapshots) == 0 {
			return fmt.Errorf("migration is active, but no snapshots found. The node might still be creating the snapshot")
		}
		if len(snapshots) > 1 {
			return fmt.Errorf("migration is active, but more than one snapshot found. This should not happen, and is likely a bug")
		}

		genesisVals := make([]*chain.GenesisValidator, len(vals))
		for i, v := range vals {
			genesisVals[i] = &chain.GenesisValidator{
				PubKey: v.PubKey,
				Power:  v.Power,
				Name:   fmt.Sprintf("validator-%d", i),
			}
		}

		go m.generateGenesisConfig(ctx, snapshots[0].SnapshotHash, genesisVals, m.Logger)
	}

	if block.Height == m.activeMigration.EndHeight {
		// starting from here, no more transactions of any kind will be accepted or mined.
		block.ChainContext.NetworkParameters.MigrationStatus = types.MigrationCompleted
		m.Logger.Info("migration to chain completed, no new transactions will be accepted")
	}

	m.lastChangeset = block.Height
	return nil
}

// generateGenesisConfig generates the genesis config for the new chain based on the snapshot and the current
// chain's consensus params. It saves the genesis file to the migrations/snapshots directory.
// This function is called only once at the start height of the migration.
// It is run asynchronously as we don't have access to the cometbft's state during replay.
// Therefore we need to wait for the consensus params fn to be set before we can generate the genesis file.
func (m *Migrator) generateGenesisConfig(ctx context.Context, snapshotHash []byte, genesisValidators []*chain.GenesisValidator, logger log.Logger) {
	// block until the m.consensusParamsFn is closed
	<-m.consensusParamsFnChan

	// sanity check
	if m.consensusParamsFn == nil {
		logger.Error("consensus params fn is nil, cannot generate genesis config")
		return
	}

	logger.Info("generating genesis config for the new chain")

	height := m.activeMigration.StartHeight - 1
	consensusParmas := m.consensusParamsFn(ctx, &height)
	if consensusParmas == nil {
		logger.Error("consensus params not found, cannot generate genesis config")
		return
	}

	var genVals []*types.NamedValidator

	for _, v := range genesisValidators {
		genVals = append(genVals, &types.NamedValidator{
			Name: v.Name,
			Validator: types.Validator{
				PubKey: v.PubKey,
				Power:  v.Power,
			},
		})
	}

	genInfo := &types.GenesisInfo{
		AppHash:    snapshotHash,
		Validators: genVals,
	}

	bts, err := json.Marshal(genInfo)
	if err != nil {
		logger.Error("failed to marshal genesis info", log.Error(err))
		return
	}

	// Save the genesis info
	err = os.WriteFile(formatGenesisInfoFileName(m.dir), bts, 0644)
	if err != nil {
		logger.Error("failed to save genesis info", log.Error(err))
		return
	}

	logger.Info("genesis config generated successfully")
}

func (m *Migrator) SetMigrationStatus(status types.MigrationStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.migrationStatus = status
}

func (m *Migrator) PersistLastChangesetHeight(ctx context.Context, tx sql.Executor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return setLastStoredChangeset(ctx, tx, m.lastChangeset)
}

// GetMigrationMetadata gets the metadata for the genesis snapshot,
// as well as the available changesets.
func (m *Migrator) GetMigrationMetadata(ctx context.Context) (*types.MigrationMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.migrationStatus == types.GenesisMigration {
		return &types.MigrationMetadata{
			MigrationState: types.MigrationState{
				Status:      types.GenesisMigration,
				StartHeight: m.genesisMigrationParams.StartHeight,
				EndHeight:   m.genesisMigrationParams.EndHeight,
			},
		}, nil
	}

	// if there is no planned migration, return
	if m.migrationStatus == types.NoActiveMigration {
		metadata := &types.MigrationMetadata{
			MigrationState: types.MigrationState{
				Status: types.NoActiveMigration,
			},
			Version: MigrationVersion,
		}

		if m.genesisMigrationParams.StartHeight != 0 && m.genesisMigrationParams.EndHeight != 0 {
			metadata.MigrationState.StartHeight = m.genesisMigrationParams.StartHeight
			metadata.MigrationState.EndHeight = m.genesisMigrationParams.EndHeight
		}
		return metadata, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Migration is triggered but not yet started
	if m.migrationStatus == types.ActivationPeriod {
		return &types.MigrationMetadata{
			MigrationState: types.MigrationState{
				Status:      types.ActivationPeriod,
				StartHeight: m.activeMigration.StartHeight,
				EndHeight:   m.activeMigration.EndHeight,
			},
			Version: MigrationVersion,
		}, nil
	}

	// Migration is in progress, retrieve the snapshot and the genesis config
	snapshots := m.snapshotter.ListSnapshots()
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("migration is active, but no snapshots found. The node might still be creating the snapshot")
	}
	if len(snapshots) > 1 {
		return nil, fmt.Errorf("migration is active, but more than one snapshot found. This should not happen, and is likely a bug")
	}

	// serialize the snapshot metadata
	snapshotBts, err := json.Marshal(snapshots[0])
	if err != nil {
		return nil, err
	}

	// read the genesis config
	genCfg, err := os.ReadFile(formatGenesisInfoFileName(m.dir))
	if err != nil {
		return nil, err
	}

	// unmarshal
	var genesisInfo types.GenesisInfo
	if err := json.Unmarshal(genCfg, &genesisInfo); err != nil {
		return nil, err
	}

	return &types.MigrationMetadata{
		MigrationState: types.MigrationState{
			Status:      m.migrationStatus,
			StartHeight: m.activeMigration.StartHeight,
			EndHeight:   m.activeMigration.EndHeight,
		},
		SnapshotMetadata: snapshotBts,
		GenesisInfo:      &genesisInfo,
		Version:          MigrationVersion,
	}, nil
}

// GetGenesisSnapshotChunk gets the snapshot chunk of Index at the given height.
func (m *Migrator) GetGenesisSnapshotChunk(chunkIdx uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeMigration == nil {
		return nil, ErrNoActiveMigration
	}

	return m.snapshotter.LoadSnapshotChunk(uint64(m.activeMigration.StartHeight), 0, chunkIdx)
}

// GetChangesetMetadata gets the metadata for the changeset at the given height.
func (m *Migrator) GetChangesetMetadata(height int64) (*ChangesetMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeMigration == nil {
		return nil, ErrNoActiveMigration
	}
	if height < m.activeMigration.StartHeight {
		return nil, fmt.Errorf("requested changeset height is before the start of the migration")
	}
	if height > m.activeMigration.EndHeight {
		return nil, fmt.Errorf("requested changeset height is after the end of the migration")
	}
	if height > m.lastChangeset {
		return nil, fmt.Errorf("requested changeset height has not been recorded by the node yet")
	}

	return loadChangesetMetadata(formatChangesetMetadataFilename(m.dir, height))
}

// GetChangeset gets the changeset at the given height and index.
func (m *Migrator) GetChangeset(height int64, index int64) ([]byte, error) {
	metadata, err := m.GetChangesetMetadata(height)
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= metadata.Chunks {
		return nil, fmt.Errorf("requested changeset index is out of bounds")
	}

	return m.loadChangeset(height, index)
}

// Directory structure:
// migrations/
//	changesets/
//	  block-1/
//	  		metadata.json [#chunks, overallsz, hash?]
//	  		chunks/
//	  			chunk-1
//	  			chunk-2
//	  			...
//	  block-2/
//	  		metadata.json [#chunks, overallsz, hash?]
//	  		chunks/
//	  			chunk-1
//	snapshots/
//		genesis.json
//		snapshot data.....

type ChangesetMetadata struct {
	Height     int64
	Chunks     int64
	ChunkSizes []int64
}

// Serialize serializes the metadata to a file.
func (m *ChangesetMetadata) saveAs(file string) error {
	bts, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, bts, 0644)
}

// LoadChangesetMetadata loads the metadata associated with a changeset.s
// It reads the changeset metadata file and returns the metadata.
func loadChangesetMetadata(metadatafile string) (*ChangesetMetadata, error) {
	bts, err := os.ReadFile(metadatafile)
	if err != nil {
		return nil, fmt.Errorf("metadata file not found %s", err.Error())
	}

	var metadata ChangesetMetadata
	if err := json.Unmarshal(bts, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

type BlockSpends struct {
	Spends []*txapp.Spend
}

var _ pg.ChangeStreamer = (*BlockSpends)(nil)

func (bs *BlockSpends) Prefix() byte {
	return pg.BlockSpendsType
}

func (bs *BlockSpends) MarshalBinary() ([]byte, error) {
	return serialize.Encode(bs)
}

var _ encoding.BinaryUnmarshaler = (*BlockSpends)(nil)

func (bs *BlockSpends) UnmarshalBinary(bts []byte) error {
	return serialize.Decode(bts, bs)
}

type chunkWriter struct {
	dir      string
	height   int64
	chunkIdx int64 // current chunk index, zero based index

	chunkFile         *os.File // current chunk file being written to
	chunkSize         int64    // number of bytes written to the current chunk file
	totalBytesWritten int64    // total number of bytes written to disk for the changeset
	chunkSizes        []int64  // sizes of each chunk
}

func newChunkWriter(dir string, height int64) *chunkWriter {
	return &chunkWriter{
		dir:      dir,
		height:   height,
		chunkIdx: -1,
	}
}

var _ io.Writer = (*chunkWriter)(nil)

func (cw *chunkWriter) Write(bts []byte) (int, error) {
	if len(bts) == 0 {
		return 0, nil
	}

	if cw.chunkFile == nil {
		cw.chunkIdx++
		filename := formatChangesetFilename(cw.dir, cw.height, cw.chunkIdx)
		file, err := os.Create(filename)
		if err != nil {
			return 0, err
		}
		cw.chunkFile = file
		cw.chunkSizes = append(cw.chunkSizes, 0)
	}

	// write the data to the file, the file can only hold a maximum of ChunkSize bytes
	maxBytesToWrite := MaxChunkSize - cw.chunkSize
	if int64(len(bts)) >= maxBytesToWrite {
		// write the maximum number of bytes that can be written to the file
		n, err := cw.chunkFile.Write(bts[:maxBytesToWrite])
		if err != nil {
			return 0, err
		}
		cw.totalBytesWritten += int64(n)
		cw.chunkSizes[cw.chunkIdx] += int64(n)

		// close the current file
		err = cw.chunkFile.Close()
		if err != nil {
			return 0, err
		}

		// increment the chunk index
		// cw.chunkIdx++
		cw.chunkSize = 0
		cw.chunkFile = nil

		// write the remaining bytes to the next file
		nRem, err := cw.Write(bts[maxBytesToWrite:])
		if err != nil {
			return 0, err
		}
		return n + nRem, nil
	}

	// write the data to the file
	n, err := cw.chunkFile.Write(bts)
	if err != nil {
		return 0, err
	}
	cw.chunkSize += int64(n)
	cw.totalBytesWritten += int64(n)
	cw.chunkSizes[cw.chunkIdx] += int64(n)
	return n, nil
}

func (cw *chunkWriter) Close() error {
	if cw.chunkFile != nil {
		return cw.chunkFile.Close()
	}
	return nil
}

func (cw *chunkWriter) SaveMetadata() error {
	filename := formatChangesetMetadataFilename(cw.dir, cw.height)
	metadata := &ChangesetMetadata{
		Height:     cw.height,
		Chunks:     cw.chunkIdx + 1,
		ChunkSizes: cw.chunkSizes,
	}
	return metadata.saveAs(filename)
}

// storeChangeset persists a changeset to the migrations/changesets directory.
// errChan is a channel that receives errors from the changeset storage routine and signals
// abci that the changeset storage has completed or failed.
func (m *Migrator) StoreChangesets(height int64, changes <-chan any, errChan chan<- error) {
	if changes == nil {
		// no changesets to store, not in a migration
		return
	}

	err := ensureChangesetDir(m.dir, height)
	if err != nil {
		errChan <- err
		return
	}

	// create a chunk writer
	chunkWriter := newChunkWriter(m.dir, height)
	defer chunkWriter.Close()

	for ch := range changes {
		switch ct := ch.(type) {
		case *pg.ChangesetEntry:
			// write the changeset to disk
			err = pg.StreamElement(chunkWriter, ct)
			if err != nil {
				errChan <- err
				return
			}

		case *pg.Relation:
			// write the relation to disk
			err = pg.StreamElement(chunkWriter, ct)
			if err != nil {
				errChan <- err
				return
			}
		}
	}

	// write the block spends to disk
	bs := &BlockSpends{
		Spends: m.accounts.GetBlockSpends(),
	}
	if len(bs.Spends) > 0 {
		if pg.StreamElement(chunkWriter, bs); err != nil {
			errChan <- err
			return
		}
	}

	if err = chunkWriter.SaveMetadata(); err != nil {
		errChan <- err
		return
	}

	// signals NotifyHeight that all changesets have been written to disk
	errChan <- nil
}

// LoadChangesets loads changesets at a given height from the migration directory.
func (m *Migrator) loadChangeset(height int64, index int64) ([]byte, error) {
	file, err := os.Open(formatChangesetFilename(m.dir, height, index))
	if err != nil {
		// we don't have to have special checks for non-existence, since
		// we should check that prior to calling this function.
		return nil, err
	}

	bts, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return bts, nil
}

const (
	changesetsDirName = "changesets"
	chunksDirName     = "chunks"
)

// ChangesetsDir returns the directory where changesets are stored,
// relative to the migration directory.
func ChangesetsDir(migrationDir string) string {
	return filepath.Join(migrationDir, changesetsDirName)
}

// ensureChangesetDir creates the directory structure for a changeset block
// if it does not exist.
// dir created: migrations/changesets/block-<height>/chunks
func ensureChangesetDir(dir string, height int64) error {
	return os.MkdirAll(formatChangsetBlockDir(dir, height), 0755)
}

func formatChangesetFilename(mdir string, height int64, index int64) string {
	chunkDir := formatChangsetBlockDir(mdir, height)
	return filepath.Join(chunkDir, fmt.Sprintf("changeset-%d", index))
}

func formatChangsetBlockDir(mdir string, height int64) string {
	return filepath.Join(mdir, changesetsDirName, fmt.Sprintf("block-%d", height), chunksDirName)
}

func formatChangesetMetadataFilename(mdir string, height int64) string {
	return filepath.Join(mdir, changesetsDirName, fmt.Sprintf("block-%d", height), metadataFileName)
}

func formatGenesisInfoFileName(mdir string) string {
	return filepath.Join(mdir, "genesis_info.json")
}

// CleanupResolutionsAtStartup is called at startup to clean up the resolutions table. It does the below things:
// - Remove all the pending migration, changeset, validator join and validator remove resolutions
// - Fix the expiry heights of all the pending resolutions
// (how to handle this for offline migrations? we have no way to know the last height of the old chain)
func CleanupResolutionsAfterMigration(ctx context.Context, db sql.DB, adjustExpiration bool, snapshotHeight int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	resolutionTypes := []string{
		voting.StartMigrationEventType,
		voting.ChangesetMigrationEventType,
		voting.ValidatorJoinEventType,
		voting.ValidatorRemoveEventType,
	}

	err = voting.DeleteResolutionsByType(ctx, tx, resolutionTypes)
	if err != nil {
		return err
	}

	if adjustExpiration {
		// Fix the expiry heights of all the pending resolutions
		err = voting.ReadjustExpirations(ctx, tx, snapshotHeight)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

type ConsensusParamsGetter func(ctx context.Context, height *int64) *cmtTypes.ConsensusParams

// SetConsensusParamsGetter sets the function that returns the consensus params for the chain.
// This closes the consensusParamsFnChan to signal that the function is set.
// This is required especially in the replay mode, where the cometbft state is not available
// until the replay is done. Therefore, the genesis config cannot be generated until the
// consensus params are available.
// SeeAlso: NewCometBftNode() in internal/abci/cometbft/node.go for the function that
// generate the node config and does the replay.
func (m *Migrator) SetConsensusParamsGetter(fn ConsensusParamsGetter) {
	m.consensusParamsFn = fn
	close(m.consensusParamsFnChan)
}
