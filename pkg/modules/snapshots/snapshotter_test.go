package snapshots_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/snapshots"
)

var (
	test_single_dbfile = &snapshots.Snapshotter{
		SnapshotDir:      "./tmp/snapshots",
		SnapshotFileType: "sqlite",
		DatabaseDir:      "./test_data/dir1/",
		ChunkSize:        64 * 1024,
		SnapshotFailed:   false,
		Snapshot:         nil,
	}

	test_multiple_dbfile = &snapshots.Snapshotter{
		SnapshotDir:      "./tmp/snapshots",
		SnapshotFileType: "sqlite",
		DatabaseDir:      "./test_data/dir2/",
		ChunkSize:        64 * 1024,
		SnapshotFailed:   false,
		Snapshot:         nil,
	}
)

func Test_Snapshot_Session_Success(t *testing.T) {
	defer cleanup()
	snapshotter := test_single_dbfile
	err := snapshotter.StartSnapshotSession(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that that directory was created
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was not created")
	}

	err = snapshotter.EndSnapshotSession()
	if err != nil {
		t.Fatal("Expected snapshot to fail")
	}

	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was supposed to be deleted")
	}

	// Metadata file should be written
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json")) {
		t.Fatal("Snapshot metadata file was not written")
	}
}

func Test_Snapshot_Session_Failure(t *testing.T) {
	defer cleanup()
	snapshotter := test_single_dbfile
	err := snapshotter.StartSnapshotSession(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that that directory was created
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was not created")
	}
	snapshotter.SnapshotFailed = true

	err = snapshotter.EndSnapshotSession()
	if err == nil {
		t.Fatal("Expected snapshot to fail")
	}

	if dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was supposed to be deleted")
	}

}

func Test_CreateSnapshot_SingleFile(t *testing.T) {
	defer cleanup()
	snapshotter := test_single_dbfile
	err := snapshotter.StartSnapshotSession(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that that directory was created
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was not created")
	}

	err = snapshotter.CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the chunks were created
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_0")) {
		t.Fatal("Snapshot chunks were not created")
	}

	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_15")) {
		t.Fatal("Snapshot chunks were not created")
	}
	err = snapshotter.EndSnapshotSession()
	if err != nil {
		t.Fatal("Expected snapshot to fail")
	}

	// Verify that the metadata file was written
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json")) {
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
	defer cleanup()
	snapshotter := test_multiple_dbfile
	err := snapshotter.StartSnapshotSession(1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify that that directory was created
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks")) {
		t.Fatal("Snapshot directory was not created")
	}

	err = snapshotter.CreateSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the chunks were created
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_0")) {
		t.Fatal("Snapshot chunks were not created")
	}

	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "chunks", "chunk_0_31")) {
		t.Fatal("Snapshot chunks were not created")
	}
	err = snapshotter.EndSnapshotSession()
	if err != nil {
		t.Fatal("Expected snapshot to fail")
	}

	// Verify that the metadata file was written
	if !dirExists(filepath.Join(snapshotter.SnapshotDir, "1", "metadata.json")) {
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

func dirExists(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func cleanup() {
	os.RemoveAll("./tmp")
}
