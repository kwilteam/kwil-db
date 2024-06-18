package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/statesync"
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
	// snapshotter creates snapshots of the state.
	snapshotter Snapshotter
	// DB is a connection to the database.
	// It should connect to the same Postgres database as kwild,
	// but should be a different connection pool.
	DB Database

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

// TODO: when we implement the constructor, we need to add a note that the snapshotter should
// not be the normal snapshotter, but instead its own instance. This is needed to ensure we do not
// delete the migration snapshot.

// NotifyHeight notifies the migrator that a new block has been committed.
// It is called at the end of the block being applied, but before the block is
// committed to the database, in between tx.PreCommit and tx.Commit.
func (m *Migrator) NotifyHeight(ctx context.Context, block *common.BlockContext) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// if there is no active migration, there is nothing to do
	if m.activeMigration == nil {
		return nil
	}

	/*
		I previously thought to make this run asynchronously, since PG dump can take a significant amount of time,
		however I decided againast it, because nodes are required to agree on the height of the old chain during the
		migration on the new chain. Im not sure of a way to guarantee this besdies literally enforcing that the old
		chain runs the migration syncrhonously as part of consensus.
	*/

	// if the current block height is the height at which the migration starts, then
	// we should snapshot the current DB and begin the migration. Since NotifyHeight is called
	// during PreCommit, the state changes from the current block won't be included in the snapshot,
	// and will instead need to be recorded as the first changeset of the migration.
	if block.Height == m.activeMigration.StartHeight {
		tx, snapshotId, err := m.DB.BeginSnapshotTx(ctx)
		if err != nil {
			return err
		}

		err = m.snapshotter.CreateSnapshot(ctx, uint64(block.Height), snapshotId)
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
	}

	if block.Height == m.activeMigration.EndHeight {
		// an error here will halt the node.
		// there might be a more elegant way to handle this, but for now, this is fine.
		return fmt.Errorf(`NETWORK HALTED: migration to chain "%s" has completed`, m.activeMigration.ChainID)
	}

	// if not in a migration, we can return early
	if block.Height < m.activeMigration.StartHeight {
		return nil
	}

	if block.Height > m.activeMigration.EndHeight {
		panic("internal bug: block height is greater than end height of migration")
	}

	// if we reach here, we are in a block that must be migrated.
	// TODO: get changeset
	var cs Changeset
	err := m.storeChangeset(&cs)
	if err != nil {
		return err
	}

	m.lastChangeset = block.Height
	return nil
}

var ErrNoActiveMigration = fmt.Errorf("no active migration")

// MigrationMetadata holds metadata about a migration, informing
// consumers of what information the current node has available
// for the migration.
type MigrationMetadata struct {
	// GenesisSnapshot holds information about the genesis snapshot.
	GenesisSnapshot *statesync.Snapshot
	// LastChangeset is the height of the last changeset that was stored.
	// Nodes are expected to have all changesets from LastChangeset to
	// Snapshot.Height. If LastChangeset is -1, then no changesets have
	// been stored yet.
	LastChangeset int64
}

// GetMigrationMetadata gets the metadata for the genesis snapshot,
// as well as the available changesets.
func (m *Migrator) GetMigrationMetadata() (*MigrationMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeMigration == nil {
		return nil, ErrNoActiveMigration
	}

	snapshots := m.snapshotter.ListSnapshots()
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("migration is active, but no snapshots found. The node might still be creating the snapshot")
	}
	if len(snapshots) > 1 {
		return nil, fmt.Errorf("migration is active, but more than one snapshot found. This should not happen, and is likely a bug")
	}

	return &MigrationMetadata{
		GenesisSnapshot: snapshots[0],
		LastChangeset:   m.lastChangeset,
	}, nil
}

// GetGenesisChunk gets the genesis chunk for the migration.
func (m *Migrator) GetGenesisChunk(height int64, format uint32, chunkIdx uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeMigration == nil {
		return nil, ErrNoActiveMigration
	}
	return m.snapshotter.LoadSnapshotChunk(uint64(height), format, chunkIdx)
}

// GetChangeset gets the changeset for a block in the migration.
func (m *Migrator) GetChangeset(height int64) (*Changeset, error) {
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

	return m.loadChangeset(height)
}

// storeChangeset persists a changeset to the migration directory.
func (m *Migrator) storeChangeset(c *Changeset) error {
	bts, err := c.MarshalBinary()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(m.dir, formatChangesetFilename(c.Height)))
	if err != nil {
		return err
	}

	_, err = file.Write(bts)
	if err != nil {
		return err
	}

	return file.Close()
}

// loadChangeset loads a changeset from the migration directory.
func (m *Migrator) loadChangeset(height int64) (*Changeset, error) {
	file, err := os.Open(filepath.Join(m.dir, formatChangesetFilename(height)))
	if err != nil {
		// we don't have to have special checks for non-existence, since
		// we should check that prior to calling this function.
		return nil, err
	}

	var c Changeset
	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

type Changeset struct {
	Height int64 `json:"height"`
}

func (c *Changeset) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

func (c *Changeset) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func formatChangesetFilename(height int64) string {
	return fmt.Sprintf("changeset-%d.json", height)
}
