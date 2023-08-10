package snapshots

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/utils"
)

type Snapshotter struct {
	SnapshotDir      string
	SnapshotFileType string
	DatabaseDir      string
	ChunkSize        uint64
	SnapshotFailed   bool
	Snapshot         *Snapshot
}

func NewSnapshotter(snapshotDir string, databaseDir string, snapshotFileType string) *Snapshotter {
	return &Snapshotter{
		SnapshotDir:      snapshotDir,
		SnapshotFileType: snapshotFileType,
		DatabaseDir:      databaseDir,
		Snapshot:         nil,
		ChunkSize:        16 * 1024 * 1024,
		SnapshotFailed:   false,
	}
}

// Creates /snapshots/<height>/chunks directory & initializes snapshot metadata
func (s *Snapshotter) StartSnapshotSession(height uint64) error {
	snapshotsPath := filepath.Join(s.SnapshotDir, fmt.Sprintf("%d", height), "chunks")
	err := os.MkdirAll(snapshotsPath, 0755)
	if err != nil {
		return err
	}

	s.Snapshot = &Snapshot{
		Height:     height,
		Format:     0,
		ChunkCount: 0,
		Hash:       nil,
		Metadata: SnapshotMetadata{
			ChunkHashes: make(map[uint32][]byte),
			FileInfo:    make(map[string]SnapshotFileInfo),
		},
	}

	return nil
}

// Writes Snapshot metadata to disk
func (s *Snapshotter) EndSnapshotSession() error {
	if s.SnapshotFailed {
		snapshotDir := filepath.Join(s.SnapshotDir, fmt.Sprintf("%d", s.Snapshot.Height))
		s.deleteSnapshot(snapshotDir)
		return fmt.Errorf("Snapshot failed")
	}

	err := s.writeSnapshotFile()
	if err != nil {
		return err
	}
	s.Snapshot = nil
	return nil
}

/*
List all the files in the database directory in sorted order (for ordering chunks)
Divide each file into chunks of 16 MB max
Hash each chunk and the entire file
Store chunk mapping to file in the snapshot metadata for restoring the DB from chunks
*/
func (s *Snapshotter) CreateSnapshot() error {
	var wg sync.WaitGroup
	startIdx := uint32(0)

	filesToSnapshot, err := s.listFilesAlphbetically(s.DatabaseDir + "/*")
	if err != nil {
		s.SnapshotFailed = true
		return err
	}

	for _, file := range filesToSnapshot {
		_, num_chunks, err := s.numChunks(file)
		if err != nil {
			s.SnapshotFailed = true
			return err
		}
		wg.Add(1)

		go func(file string, startIdx uint32, num_chunks uint32) {
			defer wg.Done()
			err := s.createFileSnaphshot(file, startIdx)
			if err != nil {
				s.SnapshotFailed = true
			}

		}(file, startIdx, num_chunks)
		startIdx += num_chunks
	}

	wg.Wait()
	if s.SnapshotFailed {
		return fmt.Errorf("Snapshot failed")
	}

	return nil
}

func (s *Snapshotter) createFileSnaphshot(file string, startIdx uint32) error {
	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()

	sz, chunks, err := s.numChunks(file)
	if err != nil {
		return err
	}

	for idx := startIdx; idx < startIdx+chunks; idx += 1 {
		err = s.createChunk(reader, idx)
		if err != nil {
			return err
		}
	}

	hash, err := utils.HashFile(file)
	if err != nil {
		return err
	}

	s.Snapshot.ChunkCount += chunks
	s.Snapshot.Metadata.FileInfo[file] = SnapshotFileInfo{
		Size:     sz,
		Hash:     hash,
		BeginIdx: startIdx,
		EndIdx:   startIdx + chunks - 1,
	}
	return nil
}

func (s *Snapshotter) createChunk(reader *os.File, chunkIdx uint32) error {
	writer, err := s.CreateChunkFile(chunkIdx)
	if err != nil {
		return err
	}
	defer writer.Close()

	chunker := NewChunker(reader, writer, int64(s.ChunkSize))
	err = chunker.chunkFile()
	if err != nil {
		return err
	}

	hash, err := utils.HashFile(writer.Name())
	if err != nil {
		return err
	}
	s.Snapshot.Metadata.ChunkHashes[chunkIdx] = hash
	return nil
}

func (s *Snapshotter) ListSnapshots() ([]Snapshot, error) {
	pathRegex := filepath.Join(s.SnapshotDir, "*", "metadata.json")
	snapshotFiles, err := filepath.Glob(pathRegex)
	if err != nil {
		return nil, err
	}

	var snapshots []Snapshot
	for _, snapshotFile := range snapshotFiles {
		snapshot, err := s.readSnapshotFile(snapshotFile)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, *snapshot)
	}

	return snapshots, nil
}

func (s *Snapshotter) LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) ([]byte, error) {
	chunkFile := s.chunkFilePath(chunkID)
	chunk, err := utils.ReadFile(chunkFile)
	if err != nil {
		return nil, err
	}
	return chunk, nil
}

func (s *Snapshotter) DeleteOldestSnapshot() error {
	oldestSnapshotDir, err := s.oldestSnapshotDir()
	if err != nil {
		return err
	}

	err = s.deleteSnapshot(oldestSnapshotDir)
	return err
}

func (s *Snapshotter) deleteSnapshot(dir string) error {
	s.Snapshot = nil
	return os.RemoveAll(dir)
}
