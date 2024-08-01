package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/chain"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
	"github.com/kwilteam/kwil-db/internal/statesync"
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
	genesisFileName  = "genesis.json"
)

const (
	ChunkSize = 4 * 1000 * 1000 // around 4MB
)

// migrator is responsible for managing the migrations.
// It is responsible for tracking any in-process migrations, snapshotting
// at the appropriate height, persisting changesets for the migration for each
// block as it occurs, and making that data available via RPC for the new node.
// Similarly, if the local process is the new node, it is responsible for reading
// changesets from the external node and applying them to the local database.
type Migrator struct {
	// mu is the mutex for the migrator.
	mu sync.RWMutex

	// activeMigration is the migration that is currently in progress.
	// It is nil if there is no migration in progress.
	activeMigration *activeMigration

	// Set to true if a migration is in progress.
	// i.e if the block height is between the start and end height of the migration.
	inProgress bool

	// snapshotter creates snapshots of the state.
	snapshotter Snapshotter

	// DB is a connection to the database.
	// It should connect to the same Postgres database as kwild,
	// but should be a different connection pool.
	DB Database

	accounts SpendTracker

	// lastChangeset is the height of the last changeset that was stored.
	// If no changesets have been stored, it is -1.
	lastChangeset int64

	// Logger is the logger for the migrator.
	Logger log.Logger

	// dir is the directory where the migration data is stored.
	// It is expected to be a full path.
	dir string
}

// activeMigration is an in-process migration.
type activeMigration struct {
	// StartHeight is the height at which the migration starts.
	StartHeight int64
	// EndHeight is the height at which the migration ends.
	EndHeight int64
	// ChainID is the chain ID of the migration.
	ChainID string
}

func SetupMigrator(ctx context.Context, db Database, snapshotter Snapshotter, accounts SpendTracker, dir string, logger log.Logger) (*Migrator, error) {
	// Set the migrator declared in migrations.go
	migrator.snapshotter = snapshotter
	migrator.Logger = logger
	migrator.dir = dir
	migrator.DB = db
	migrator.accounts = accounts
	migrator.lastChangeset = -1

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

	return migrator, nil
}

// NotifyHeight notifies the migrator that a new block has been committed.
// It is called at the end of the block being applied, but before the block is
// committed to the database, in between tx.PreCommit and tx.Commit.
func (m *Migrator) NotifyHeight(ctx context.Context, block *common.BlockContext, db Database, csReader io.Reader) error {
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
		m.inProgress = false
		panic("internal bug: block height is greater than end height of migration")
	}

	if block.Height == m.activeMigration.StartHeight-1 {
		// set the migration in progress, so that we record the changesets starting from the next block
		m.inProgress = true
		block.ChainContext.NetworkParameters.InMigration = true
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

		err = m.snapshotter.CreateSnapshot(ctx, uint64(block.Height), snapshotId, networkMigrationSchemas, networkMigrationExcludedTables, networkMigrationExcludedTableData)
		if err != nil {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				// we can mostly ignore this error, since the original err will halt the node anyways
				m.Logger.Errorf("failed to rollback transaction: %s", err2.Error())
			}
			return err
		}

		// Generate a genesis file for the snapshot
		vals, err := voting.GetValidators(ctx, tx)
		if err != nil {
			err2 := tx.Rollback(ctx)
			if err2 != nil {
				// we can mostly ignore this error, since the original err will halt the node anyways
				m.Logger.Errorf("failed to rollback transaction: %s", err2.Error())
			}
			return err
		}

		err = tx.Rollback(ctx)
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

		genCfg := chain.DefaultGenesisConfig()
		genesisVals := make([]*chain.GenesisValidator, len(vals))
		for i, v := range vals {
			genesisVals[i] = &chain.GenesisValidator{
				PubKey: v.PubKey,
				Power:  v.Power,
				Name:   fmt.Sprintf("validator-%d", i),
			}
		}

		genCfg.Validators = genesisVals
		genCfg.DataAppHash = snapshots[0].SnapshotHash
		genCfg.ChainID = m.activeMigration.ChainID

		// Save the genesis file
		err = genCfg.SaveAs(formatGenesisFilename(m.dir))
		if err != nil {
			return err
		}
	}

	if block.Height == m.activeMigration.EndHeight {
		// an error here will halt the node.
		// there might be a more elegant way to handle this, but for now, this is fine.
		return fmt.Errorf(`NETWORK HALTED: migration to chain "%s" has completed`, m.activeMigration.ChainID)
	}

	// if we reach here, we are in a block that must be migrated.
	err := m.storeChangesets(block.Height, csReader)
	if err != nil {
		return err
	}

	m.lastChangeset = block.Height
	return nil
}

// MigrationMetadata holds metadata about a migration, informing
// consumers of what information the current node has available
// for the migration.
type MigrationMetadata struct {
	// InMigration is true if the node is currently in a migration process.
	// block height > StartHeight
	InMigration bool

	// GenesisSnapshot holds information about the genesis snapshot.
	GenesisSnapshot *statesync.Snapshot

	// GenesisConfig is the genesis configuration file data for the migration.
	// It is the configuration that the new chain should use.
	GenesisConfig *chain.GenesisConfig

	// LastChangeset is the height of the last changeset that was stored.
	// Nodes are expected to have all changesets from LastChangeset to
	// Snapshot.Height. If LastChangeset is -1, then no changesets have
	// been stored yet.
	LastChangeset int64

	// StartHeight is the height at which the migration starts.
	StartHeight int64

	// EndHeight is the height at which the migration ends.
	EndHeight int64
}

func (mm *MigrationMetadata) MarshalBinary() ([]byte, error) {
	return json.Marshal(mm)
}

func (mm *MigrationMetadata) UnmarshalBinary(bts []byte) error {
	return json.Unmarshal(bts, mm)
}

// GetMigrationMetadata gets the metadata for the genesis snapshot,
// as well as the available changesets.
func (m *Migrator) GetMigrationMetadata() (*MigrationMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// if there is no planned migration, return
	if m.activeMigration == nil {
		return nil, ErrNoActiveMigration
	}

	// Migration is triggered but not yet started
	if !m.inProgress {
		return &MigrationMetadata{
			InMigration: false,
			StartHeight: m.activeMigration.StartHeight,
			EndHeight:   m.activeMigration.EndHeight,
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

	genCfg, err := chain.LoadGenesisConfig(formatGenesisFilename(m.dir))
	if err != nil {
		return nil, err
	}

	return &MigrationMetadata{
		InMigration:     true,
		GenesisSnapshot: snapshots[0],
		GenesisConfig:   genCfg,
		LastChangeset:   m.lastChangeset,
		StartHeight:     m.activeMigration.StartHeight,
		EndHeight:       m.activeMigration.EndHeight,
	}, nil
}

// GetGenesisSnapshotChunk gets the snapshot chunk of Index at the given height.
func (m *Migrator) GetGenesisSnapshotChunk(height int64, format uint32, chunkIdx uint32) ([]byte, error) {
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

	return m.snapshotter.LoadSnapshotChunk(uint64(height), format, chunkIdx)
}

// GetChangesetMetadata gets the metadata for the changeset at the given height.
func (m *Migrator) GetChangesetMetadata(height int64) (*ChangesetMetdata, error) {
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

	if index < 1 || index > metadata.Chunks {
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

type ChangesetMetdata struct {
	Height        int64
	Chunks        int64
	ChangesetSize int64
	// Hash          [][HashLen]byte
}

// Serialize serializes the metadata to a file.
func (m *ChangesetMetdata) saveAs(file string) error {
	bts, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, bts, 0644)
}

// LoadChangesetMetadata loads the metadata associated with a changeset.
// It reads the changeset metadata file and returns the metadata.
func loadChangesetMetadata(metadatafile string) (*ChangesetMetdata, error) {
	bts, err := os.ReadFile(metadatafile)
	if err != nil {
		return nil, err
	}

	var metadata ChangesetMetdata
	if err := json.Unmarshal(bts, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

type BlockChangesets struct {
	Changesets []*pg.Changeset
	Spends     []*txapp.Spend
}

func (b *BlockChangesets) MarshalBinary() ([]byte, error) {
	return serialize.Encode(b)
}

func (b *BlockChangesets) UnmarshalBinary(bts []byte) error {
	return serialize.Decode(bts, b)
}

// storeChangeset persists a changeset to the migrations/changesets directory.
func (m *Migrator) storeChangesets(height int64, csReader io.Reader) error {
	if csReader == nil {
		// no changesets to store, since we are not in a migration
		return nil
	}

	// Deserialize the changesets
	cs, err := pg.DeserializeChangeset(csReader)
	if err != nil {
		return err
	}

	var filteredChangesets []*pg.Changeset

	// filter changesets to only include the changes in the user deployed schemas
	for _, changeset := range cs.Changesets {
		// changeset.Schema should start with ds_ to be included
		if strings.HasPrefix(changeset.Schema, "ds_") {
			filteredChangesets = append(filteredChangesets, changeset)
			continue
		}
	}

	blockChangesets := &BlockChangesets{
		Changesets: filteredChangesets,
		Spends:     m.accounts.GetBlockSpends(),
	}

	bts, err := blockChangesets.MarshalBinary()
	if err != nil {
		return err
	}

	// ensure the changeset directory exists
	err = ensureChangesetDir(m.dir, height)
	if err != nil {
		return err
	}

	idx := int64(0)
	// split the changeset into chunks and save them to disk
	for startIdx := 0; startIdx < len(bts); startIdx += ChunkSize {
		endIdx := startIdx + ChunkSize
		idx++
		if endIdx > len(bts) {
			endIdx = len(bts)
		}

		chunkFilename := formatChangesetFilename(m.dir, height, idx)
		err = os.WriteFile(chunkFilename, bts[startIdx:endIdx], 0644)
		if err != nil {
			return err
		}

	}

	// save the metadata for the changeset
	metadata := &ChangesetMetdata{
		Height:        height,
		Chunks:        idx,
		ChangesetSize: int64(len(bts)),
	}
	return metadata.saveAs(formatChangesetMetadataFilename(m.dir, height))
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

// ensureChangesetDir creates the directory structure for a changeset block
// if it does not exist.
// dir created: migrations/changesets/block-<height>/chunks
func ensureChangesetDir(dir string, height int64) error {
	return os.MkdirAll(formatChangsetBlockDir(dir, height), 0755)
}

func formatChangesetFilename(mdir string, height int64, index int64) string {
	chunkDir := formatChangsetBlockDir(mdir, height)
	return filepath.Join(chunkDir, fmt.Sprintf("changeset-%d.json", index))
}

func formatChangsetBlockDir(mdir string, height int64) string {
	return filepath.Join(mdir, config.ChangesetsDirName, fmt.Sprintf("block-%d", height), config.ChunksDirName)

}

func formatChangesetMetadataFilename(mdir string, height int64) string {
	return filepath.Join(mdir, config.ChangesetsDirName, fmt.Sprintf("block-%d", height), metadataFileName)
}

func formatGenesisFilename(mdir string) string {
	return filepath.Join(mdir, config.SnapshotDirName, genesisFileName)
}
