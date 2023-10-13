package snapshots_test

import (
	"crypto/sha256"
	"path/filepath"
	"testing"

	snapPkg "github.com/kwilteam/kwil-db/internal/abci/snapshots"
	"github.com/kwilteam/kwil-db/internal/utils"

	"github.com/stretchr/testify/assert"
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
	tempDir := t.TempDir()
	dbDir := filepath.Join(tempDir, "db")
	rcvdSnaps := filepath.Join(tempDir, "rcvdSnaps")
	bootstrapper, err := snapPkg.NewBootstrapper(dbDir, rcvdSnaps)
	assert.NoError(t, err)
	utils.CreateDirIfNeeded(dbDir)
	err = bootstrapper.OfferSnapshot(ValidSnapshot)
	assert.NoError(t, err)

	InvalidChunk := []byte("InvalidChunk")
	idx, status, err := bootstrapper.ApplySnapshotChunk(InvalidChunk, 0)
	assert.Error(t, err)
	assert.Equal(t, status, snapPkg.RETRY, "Invalid chunk accepted")
	assert.NotEqual(t, len(idx), 0, "Expected refetch indexes")

	InvalidChunk2 := []byte("CK_BEG::sample chunk dataCK_END")
	idx, status, err = bootstrapper.ApplySnapshotChunk(InvalidChunk2, 0)
	assert.Error(t, err)
	assert.Equal(t, status, snapPkg.RETRY, "Invalid chunk accepted")
	assert.NotEqual(t, len(idx), 0, "Expected refetch indexes")

	db_restored := bootstrapper.IsDBRestored()
	assert.False(t, db_restored, "Expected DB to not be restored")

	// Verify that db file exists
	exits := utils.FileExists("./tmp/db/file1")
	assert.False(t, exits, "db file shouldn't exist")
}

func Test_ValidSnapshot(t *testing.T) {
	tempDir := t.TempDir()

	dbDir := filepath.Join(tempDir, "db")
	rcvdSnaps := filepath.Join(tempDir, "rcvdSnaps")
	bootstrapper, err := snapPkg.NewBootstrapper(dbDir, rcvdSnaps)
	assert.NoError(t, err)
	utils.CreateDirIfNeeded(dbDir)

	err = bootstrapper.OfferSnapshot(ValidSnapshot)
	assert.NoError(t, err)

	exists := utils.FileExists(rcvdSnaps)
	assert.True(t, exists, "Expected snapshot to be written")

	idx, status, err := bootstrapper.ApplySnapshotChunk(sampleChunk, 0)
	assert.NoError(t, err)
	assert.Equal(t, snapPkg.ACCEPT, status, "Invalid chunk accepted")
	assert.Equal(t, len(idx), 0, "Expected no refetch indexes")

	db_restored := bootstrapper.IsDBRestored()
	assert.True(t, db_restored, "Expected DB to be restored")

	// Verify that db file exists
	exists = utils.FileExists(filepath.Join(dbDir, "file1"))
	assert.True(t, exists, "Expected db file to exist")
}

func Test_InValidSnapshot(t *testing.T) {
	tempDir := t.TempDir()
	dbDir := filepath.Join(tempDir, "db")
	rcvdSnaps := filepath.Join(tempDir, "rcvdSnaps")
	bootstrapper, err := snapPkg.NewBootstrapper(dbDir, rcvdSnaps)
	assert.NoError(t, err)
	utils.CreateDirIfNeeded(dbDir)

	err = bootstrapper.OfferSnapshot(InvalidSnapshot)
	assert.Nil(t, err)

	_, status, err := bootstrapper.ApplySnapshotChunk(sampleChunk, 0)
	assert.Error(t, err)
	assert.Equal(t, snapPkg.REJECT, status, "Invalid chunk accepted")

	db_restored := bootstrapper.IsDBRestored()
	assert.False(t, db_restored, "Expected DB to not be restored")

	// Verify that db file exists
	exists := utils.FileExists(filepath.Join(dbDir, "file1"))
	assert.False(t, exists, "Expected db file to exist")
}
