package migrations

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
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
	genesisFileName  = "genesis.json"
)

const (
	MaxChunkSize = 4 * 1000 * 1000 // around 4MB
)

// migrator is responsible for managing the migrations.
// It is responsible for tracking any in-process migrations, snapshotting
// at the appropriate height, persisting changesets for the migration for each
// block as it occurs, and making that data available via RPC for the new node.
// Similarly, if the local process is the new node, it is responsible for reading
// changesets from the external node and applying them   to the local database.
type Migrator struct {
	initialized bool // set to true after the migrator is initialized

	// mu is the mutex for the migrator.
	mu sync.RWMutex

	// activeMigration is the migration plan that is approved by the network.
	// It is nil if there is no plan for a migration.
	activeMigration *activeMigration

	// Set to true when the migration is in progress.
	// i.e the block height is between the start and end height of the migration.
	inProgress bool

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

	// doneChan is a channel that is closed when all the block changes have been written to disk.
	doneChan chan bool

	// errChan is a channel that receives errors from the changeset storage routine.
	errChan chan error
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

// SetupMigrator initializes the migrator instance with the necessary dependencies.
func SetupMigrator(ctx context.Context, db Database, snapshotter Snapshotter, accounts SpendTracker, dir string, logger log.Logger) (*Migrator, error) {
	if migrator.initialized {
		return nil, fmt.Errorf("migrator already initialized")
	}

	// Set the migrator declared in migrations.go
	migrator.snapshotter = snapshotter
	migrator.Logger = logger
	migrator.dir = dir
	migrator.DB = db
	migrator.accounts = accounts
	migrator.doneChan = make(chan bool, 1)
	migrator.initialized = true

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
		m.inProgress = false
		panic("internal bug: block height is greater than end height of migration")
	}

	if block.Height == m.activeMigration.StartHeight-1 {
		// set the migration in progress, so that we record the changesets starting from the next block
		m.inProgress = true
		block.ChainContext.NetworkParameters.InMigration = true
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

		genCfg.ConsensusParams.Migration.StartHeight = m.activeMigration.StartHeight
		genCfg.ConsensusParams.Migration.EndHeight = m.activeMigration.EndHeight

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

	// wait for signal on doneChan, indicating that all changesets have been written to disk
	select {
	case <-m.doneChan:
		break
	case err := <-m.errChan:
		return err
	case <-ctx.Done():
		return nil
	}

	m.lastChangeset = block.Height
	return nil
}

func (m *Migrator) PersistLastChangesetHeight(ctx context.Context, tx sql.Executor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return setLastStoredChangeset(ctx, tx, m.lastChangeset)
}

func (m *Migrator) InMigration(height int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeMigration == nil {
		return false
	}

	return height >= m.activeMigration.StartHeight
}

// GetMigrationMetadata gets the metadata for the genesis snapshot,
// as well as the available changesets.
func (m *Migrator) GetMigrationMetadata() (*types.MigrationMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// if there is no planned migration, return
	if m.activeMigration == nil {
		return nil, ErrNoActiveMigration
	}

	// Migration is triggered but not yet started
	if !m.inProgress {
		return &types.MigrationMetadata{
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

	// serialize the snapshot metadata
	snapshotBts, err := json.Marshal(snapshots[0])
	if err != nil {
		return nil, err
	}

	genCfg, err := chain.LoadGenesisConfig(formatGenesisFilename(m.dir))
	if err != nil {
		return nil, err
	}

	// serialize genesis config data
	configBts, err := json.Marshal(genCfg)
	if err != nil {
		return nil, err
	}

	return &types.MigrationMetadata{
		InMigration:      true,
		StartHeight:      m.activeMigration.StartHeight,
		EndHeight:        m.activeMigration.EndHeight,
		SnapshotMetadata: snapshotBts,
		GenesisConfig:    configBts,
	}, nil
}

// GetGenesisSnapshotChunk gets the snapshot chunk of Index at the given height.
func (m *Migrator) GetGenesisSnapshotChunk(height int64, format uint32, chunkIdx uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeMigration == nil {
		return nil, ErrNoActiveMigration
	}

	if height != m.activeMigration.StartHeight {
		return nil, fmt.Errorf("requested snapshot height is not the start of the migration")
	}

	return m.snapshotter.LoadSnapshotChunk(uint64(height), format, chunkIdx)
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

func (bs *BlockSpends) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(pg.BlockSpendsType)

	// serialize the spends
	bts, err := serialize.Encode(bs)
	if err != nil {
		return nil, err
	}

	// write the length of the spends
	size := uint32(len(bts))
	if err = binary.Write(buf, binary.LittleEndian, size); err != nil {
		return nil, err
	}

	// write the spends
	if _, err = buf.Write(bts); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (bs *BlockSpends) Deserialize(bts []byte) error {
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
		dir:    dir,
		height: height,
	}
}

func (cw *chunkWriter) Write(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}

	if cw.chunkFile == nil {
		filename := formatChangesetFilename(cw.dir, cw.height, cw.chunkIdx)
		file, err := os.Create(filename)
		if err != nil {
			return err
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
			return err
		}
		cw.totalBytesWritten += int64(n)
		cw.chunkSizes[cw.chunkIdx] += int64(n)

		// close the current file
		err = cw.chunkFile.Close()
		if err != nil {
			return err
		}

		// increment the chunk index
		cw.chunkIdx++
		cw.chunkSize = 0
		cw.chunkFile = nil

		// write the remaining bytes to the next file
		return cw.Write(bts[maxBytesToWrite:])
	}

	// write the data to the file
	n, err := cw.chunkFile.Write(bts)
	if err != nil {
		return err
	}
	cw.chunkSize += int64(n)
	cw.totalBytesWritten += int64(n)
	cw.chunkSizes[cw.chunkIdx] += int64(n)
	return nil
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
func (m *Migrator) StoreChangesets(height int64, changes <-chan any) {
	if changes == nil {
		// no changesets to store, not in a migration
		return
	}

	err := ensureChangesetDir(m.dir, height)
	if err != nil {
		m.errChan <- err
		return
	}

	// create a chunk writer
	chunkWriter := newChunkWriter(m.dir, height)
	defer chunkWriter.Close()

	for ch := range changes {
		switch ct := ch.(type) {
		case *pg.ChangesetEntry:
			// write the changeset to disk
			ce := ct

			// serialize the changeset entry and write it to disk
			bts, err := ce.Serialize()
			if err != nil {
				m.errChan <- err
				return
			}

			// Write the changeset bts to disk
			if err = chunkWriter.Write(bts); err != nil {
				m.errChan <- err
				return
			}
		case *pg.Relation:
			// write the relation to disk
			relation := ct
			// serialize the relation and write it to disk
			bts, err := relation.Serialize()
			if err != nil {
				m.errChan <- err
				return
			}

			// Write the relation bts to disk
			if err = chunkWriter.Write(bts); err != nil {
				m.errChan <- err
				return
			}
		}
	}

	// write the block spends to disk
	bs := &BlockSpends{
		Spends: m.accounts.GetBlockSpends(),
	}

	bts, err := bs.Serialize()
	if err != nil {
		m.errChan <- err
		return
	}

	if err = chunkWriter.Write(bts); err != nil {
		m.errChan <- err
		return
	}

	if err = chunkWriter.SaveMetadata(); err != nil {
		m.errChan <- err
		return
	}

	// signals NotifyHeight that all changesets have been written to disk
	m.doneChan <- true
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
	return filepath.Join(chunkDir, fmt.Sprintf("changeset-%d", index))
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
