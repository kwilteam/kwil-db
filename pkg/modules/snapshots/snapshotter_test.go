package snapshots_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/kwilteam/kwil-db/pkg/utils"
	"github.com/stretchr/testify/assert"
)

var (
	test_single_dbfile = &snapshots.Snapshotter{
		SnapshotFileType: "sqlite",
		DatabaseDir:      "./test_data/dir1/",
		ChunkSize:        64 * 1024,
		SnapshotFailed:   false,
		Snapshot:         nil,
	}

	test_multiple_dbfile = &snapshots.Snapshotter{
		SnapshotFileType: "sqlite",
		DatabaseDir:      "./test_data/dir2/",
		ChunkSize:        64 * 1024,
		SnapshotFailed:   false,
		Snapshot:         nil,
	}
)

func newTestSnapshotter(sample *snapshots.Snapshotter) *snapshots.Snapshotter {
	return &snapshots.Snapshotter{
		SnapshotDir:      sample.SnapshotDir,
		SnapshotFileType: sample.SnapshotFileType,
		DatabaseDir:      sample.DatabaseDir,
		ChunkSize:        sample.ChunkSize,
		SnapshotFailed:   false,
		Snapshot:         nil,
	}
}

func Test_Snapshot_Session_Success(t *testing.T) {
	tempDir := t.TempDir()
	fmt.Println(tempDir)
	snapshotter := newTestSnapshotter(test_single_dbfile)
	snapshotter.SnapshotDir = filepath.Join(tempDir, "snapshots")
	err := snapshotter.StartSnapshotSession(1)
	assert.NoError(t, err, "Snapshot session failed")

	// Verify that that directory was created
	exists := utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks"))
	assert.True(t, exists, "Snapshot directory was not created")

	err = snapshotter.EndSnapshotSession()
	assert.NoError(t, err, "Snapshot session failed")

	exists = utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks"))
	assert.True(t, exists, "Snapshot directory was not deleted")

	// Metadata file should be written

	exists = utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json"))
	assert.True(t, exists, "Snapshot metadata file was not written")
}

func Test_Snapshot_Session_Failure(t *testing.T) {
	tempDir := t.TempDir()
	fmt.Println(tempDir)
	snapshotter := newTestSnapshotter(test_single_dbfile)
	snapshotter.SnapshotDir = filepath.Join(tempDir, "snapshots")
	err := snapshotter.StartSnapshotSession(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that that directory was created
	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was not created")
	}
	snapshotter.SnapshotFailed = true
	snapshotter.SnapshotError = fmt.Errorf("Snapshot failed")

	err = snapshotter.EndSnapshotSession()
	if err == nil {
		t.Fatal("Expected snapshot to fail")
	}

	if utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was supposed to be deleted")
	}

}

func Test_CreateSnapshot_SingleFile(t *testing.T) {
	tempDir := t.TempDir()
	fmt.Println(tempDir)
	snapshotter := newTestSnapshotter(test_single_dbfile)
	snapshotter.SnapshotDir = filepath.Join(tempDir, "snapshots")
	err := snapshotter.StartSnapshotSession(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that that directory was created
	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was not created")
	}

	err = snapshotter.CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the chunks were created
	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_0")) {
		t.Fatal("Snapshot chunks were not created")
	}

	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_15")) {
		t.Fatal("Snapshot chunks were not created")
	}
	err = snapshotter.EndSnapshotSession()
	if err != nil {
		t.Fatal("Expected snapshot to fail")
	}

	// Verify that the metadata file was written
	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json")) {
		t.Fatal("Snapshot metadata file was not written")
	}

	metadataFile := filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json")
	snapshot, err := snapshotter.ReadSnapshotFile(metadataFile)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(snapshot)
	if snapshot.Height != 1 {
		t.Fatal("Snapshot height was not set correctly")
	}

	if snapshot.ChunkCount != 17 {
		t.Fatal("Snapshot chunk count was not set correctly")
	}

}

// This will test the concurrency of the snapshotter
func Test_CreateSnapshot_MultipleFiles(t *testing.T) {
	tempDir := t.TempDir()
	fmt.Println(tempDir)
	snapshotter := newTestSnapshotter(test_multiple_dbfile)
	snapshotter.SnapshotDir = filepath.Join(tempDir, "snapshots")
	err := snapshotter.StartSnapshotSession(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that that directory was created
	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was not created")
	}

	err = snapshotter.CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the chunks were created
	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_0")) {
		t.Fatal("Snapshot chunks were not created")
	}

	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_31")) {
		t.Fatal("Snapshot chunks were not created")
	}
	err = snapshotter.EndSnapshotSession()
	if err != nil {
		t.Fatal("Expected snapshot to fail")
	}

	// Verify that the metadata file was written
	if !utils.FileExists(filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json")) {
		t.Fatal("Snapshot metadata file was not written")
	}

	metadataFile := filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json")
	snapshot, err := snapshotter.ReadSnapshotFile(metadataFile)
	if err != nil {
		t.Fatal(err)
	}
	// fmt.Println(snapshot)
	if snapshot.Height != 1 {
		t.Fatal("Snapshot height was not set correctly")
	}

	if snapshot.ChunkCount != 34 {
		t.Fatal("Snapshot chunk count was not set correctly")
	}

	if len(snapshot.Metadata.FileInfo) != 2 {
		t.Fatal("Snapshot file info was not set correctly")
	}
}
