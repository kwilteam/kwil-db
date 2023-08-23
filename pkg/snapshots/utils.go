package snapshots

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kwilteam/kwil-db/pkg/utils"
)

func (s *Snapshotter) writeSnapshotFile() error {
	metadata, err := json.MarshalIndent(s.Snapshot, "", "  ")
	if err != nil {
		return err
	}

	metadataFilePath := filepath.Join(s.SnapshotDir, fmt.Sprintf("%d", s.Snapshot.Height), "metadata.json")
	return utils.WriteFile(metadataFilePath, metadata)
}

func (s *Snapshotter) ReadSnapshotFile(filePath string) (*Snapshot, error) {
	bts, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var snapshot Snapshot
	err = json.Unmarshal(bts, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (s *Snapshotter) listFilesAlphbetically(filePath string) ([]string, error) {
	files, err := filepath.Glob(filePath)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func (s *Snapshotter) oldestSnapshotDir() (string, error) {
	files, err := s.listFilesAlphbetically(s.SnapshotDir + "/*")
	if err != nil {
		return "", err
	}
	return files[0], nil
}

func (s *Snapshotter) chunkDir(height uint64) string {
	return filepath.Join(s.SnapshotDir, fmt.Sprintf("%d", height), "chunks")
}

func (s *Snapshotter) chunkFilePath(chunkID uint32) string {
	return filepath.Join(s.chunkDir(s.Snapshot.Height), fmt.Sprintf("chunk_%d_%d", s.Snapshot.Format, chunkID))
}

func (s *Snapshotter) numChunks(filePath string) (uint64, uint32, error) {
	fInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, 0, err
	}
	chunks := uint32(uint64(fInfo.Size()) / s.ChunkSize)
	if uint64(fInfo.Size())%s.ChunkSize != 0 {
		chunks++
	}
	return uint64(fInfo.Size()), chunks, nil
}

func (s *Snapshotter) CreateChunkFile(chunkID uint32) (*os.File, error) {
	chunkFile := s.chunkFilePath(chunkID)
	file, err := os.OpenFile(chunkFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
