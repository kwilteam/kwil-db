package snapshots_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/modules/snapshots"
	"github.com/kwilteam/kwil-db/pkg/utils"

	snapPkg "github.com/kwilteam/kwil-db/pkg/snapshots"
)

var (
	sampleChunk     = []byte("CK_BEGINsample chunk dataCK_END")
	sampleChunkHash = sha256.Sum256(sampleChunk)
	fileData        = []byte("sample chunk data")
	fileHash        = sha256.Sum256(fileData)
	invalidFileHash = sha256.Sum256([]byte("invalid file data"))
	ValidSnapshot   = &snapPkg.Snapshot{
		Height:     1,
		Format:     0,
		ChunkCount: 1,
		Metadata: snapPkg.SnapshotMetadata{
			ChunkHashes: map[uint32][]byte{
				0: sampleChunkHash[:],
			},
			FileInfo: map[string]snapPkg.SnapshotFileInfo{
				"file1": {
					Size:     uint64(len(fileData)),
					Hash:     fileHash[:],
					BeginIdx: 0,
					EndIdx:   0,
				},
			},
		},
	}

	InvalidSnapshot = &snapPkg.Snapshot{
		Height:     1,
		Format:     0,
		ChunkCount: 1,
		Metadata: snapPkg.SnapshotMetadata{
			ChunkHashes: map[uint32][]byte{
				0: sampleChunkHash[:],
			},
			FileInfo: map[string]snapPkg.SnapshotFileInfo{
				"file1": {
					Size:     uint64(len(fileData)),
					Hash:     invalidFileHash[:],
					BeginIdx: 0,
					EndIdx:   0,
				},
			},
		},
	}

	MultiFileSnapshot = &snapPkg.Snapshot{
		Height:     1,
		Format:     0,
		ChunkCount: 1,
		Metadata: snapPkg.SnapshotMetadata{
			ChunkHashes: map[uint32][]byte{
				0: sampleChunkHash[:],
				1: sampleChunkHash[:],
			},
			FileInfo: map[string]snapPkg.SnapshotFileInfo{
				"file1": {
					Size:     uint64(len(fileData)),
					Hash:     fileHash[:],
					BeginIdx: 0,
					EndIdx:   0,
				},
				"file2": {
					Size:     uint64(len(fileData)),
					Hash:     fileHash[:],
					BeginIdx: 1,
					EndIdx:   1,
				},
			},
		},
	}
)

func Test_Chunk_Validation(t *testing.T) {
	defer cleanup()
	bootstrapper := snapshots.NewBootstrapper("./tmp/rcvdsnaps/", "./tmp/db/")
	utils.CreateDirIfNeeded("./tmp/db/")
	err := bootstrapper.OfferSnapshot(ValidSnapshot)
	if err != nil {
		t.Fatal(err)
	}

	InvalidChunk := []byte("InvalidChunk")
	idx, err := bootstrapper.ApplySnapshotChunk(InvalidChunk, 0)
	fmt.Println(idx, err)
	if err == nil {
		t.Fatal("Expected error")
	}

	if len(idx) == 0 {
		t.Fatal("Expected refetch indexes")
	}

	InvalidChunk2 := []byte("CK_BEG::sample chunk dataCK_END")
	idx, err = bootstrapper.ApplySnapshotChunk(InvalidChunk2, 0)
	fmt.Println(idx, err)
	if err == nil {
		t.Fatal("Expected error")
	}

	if len(idx) == 0 {
		t.Fatal("Expected refetch indexes")
	}

	if bootstrapper.IsDBRestored() {
		t.Fatal("Expected DB to not be restored")
	}

	// Verify that db file exists
	if utils.FileExists("./tmp/db/file1") {
		t.Fatal("Expected db file to exist")
	}
}

func Test_ValidSnapshot(t *testing.T) {
	defer cleanup()
	bootstrapper := snapshots.NewBootstrapper("./tmp/rcvdsnaps/", "./tmp/db/")
	utils.CreateDirIfNeeded("./tmp/db/")
	err := bootstrapper.OfferSnapshot(ValidSnapshot)
	if err != nil {
		t.Fatal(err)
	}

	idx, err := bootstrapper.ApplySnapshotChunk(sampleChunk, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(idx) != 0 {
		t.Fatal("Expected no refetch indexes")
	}

	if !bootstrapper.IsDBRestored() {
		t.Fatal("Expected DB to be restored")
	}

	// Verify that db file exists
	if !utils.FileExists("./tmp/db/file1") {
		t.Fatal("Expected db file to exist")
	}
}

func Test_InValidSnapshot(t *testing.T) {
	defer cleanup()
	bootstrapper := snapshots.NewBootstrapper("./tmp/rcvdsnaps/", "./tmp/db/")
	utils.CreateDirIfNeeded("./tmp/db/")
	err := bootstrapper.OfferSnapshot(InvalidSnapshot)
	if err != nil {
		t.Fatal(err)
	}

	_, err = bootstrapper.ApplySnapshotChunk(sampleChunk, 0)
	if err == nil {
		t.Fatal(err)
	}

	if bootstrapper.IsDBRestored() {
		t.Fatal("Expected DB to not be restored")
	}

	// Verify that db file exists
	if utils.FileExists("./tmp/db/file1") {
		t.Fatal("Expected db file to exist")
	}
}
