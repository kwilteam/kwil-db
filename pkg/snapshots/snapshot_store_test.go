package snapshots_test

import (
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/stretchr/testify/assert"
)

func Test_SnapshotStore_Create(t *testing.T) {
	tempDir := t.TempDir()
	ss := snapshots.NewSnapshotStore("../../snapshots/test_data/dir1/",
		filepath.Join(tempDir, "snapshots"),
		1, 1)

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
