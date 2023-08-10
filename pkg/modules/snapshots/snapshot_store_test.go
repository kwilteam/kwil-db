package snapshots_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/modules/snapshots"
)

func Test_SnapshotStore_Create(t *testing.T) {
	defer cleanup()
	ss := snapshots.NewSnapshotStore(snapshots.WithEnabled(true),
		snapshots.WithSnapshotDir("./tmp/snapshots"),
		snapshots.WithDatabaseDir("./test_data/dir1/"),
		snapshots.WithDatabaseType("sqlite"),
		snapshots.WithMaxSnapshots(1),
		snapshots.WithRecurringHeight(1),
		snapshots.WithChunkSize(1*1024*1024),
		snapshots.WithSnapshotter(),
	)

	if ss == nil {
		t.Fatal("Snapshot store was not created")
	}

	err := ss.CreateSnapshot(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that snapshot store has 1 snapshot record
	numSnaps := ss.NumSnapshots()
	if numSnaps != 1 {
		t.Fatal("Snapshot store should have 1 snapshot record")
	}

	// This should delete the previous snapshot and create a new snapshot
	err = ss.CreateSnapshot(2)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that snapshot store has 1 snapshot record
	numSnaps = ss.NumSnapshots()
	if numSnaps != 1 {
		t.Fatal("Snapshot store should have 1 snapshot record")
	}
}
