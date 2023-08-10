package snapshots_test

import (
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/modules/snapshots"
	"github.com/stretchr/testify/assert"
)

func Test_SnapshotStore_Create(t *testing.T) {
	tempDir := t.TempDir()
	ss := snapshots.NewSnapshotStore(snapshots.WithEnabled(true),
		snapshots.WithSnapshotDir(filepath.Join(tempDir, "snapshots")),
		snapshots.WithDatabaseDir("./test_data/dir1/"),
		snapshots.WithDatabaseType("sqlite"),
		snapshots.WithMaxSnapshots(1),
		snapshots.WithRecurringHeight(1),
		snapshots.WithChunkSize(1*1024*1024),
		snapshots.WithSnapshotter(),
	)

	assert.NotNil(t, ss, "Snapshot store was not created")

	err := ss.CreateSnapshot(1)
	assert.NoError(t, err, "Snapshot creation failed")

	// Verify that snapshot store has 1 snapshot record
	numSnaps := ss.NumSnapshots()
	assert.Equal(t, uint64(1), numSnaps, "Snapshot store should have 1 snapshot record")

	// This should delete the previous snapshot and create a new snapshot
	err = ss.CreateSnapshot(2)
	assert.NoError(t, err, "Snapshot creation failed")

	// Verify that snapshot store has 1 snapshot record
	numSnaps = ss.NumSnapshots()
	assert.Equal(t, uint64(1), numSnaps, "Snapshot store should have 1 snapshot record")
}
