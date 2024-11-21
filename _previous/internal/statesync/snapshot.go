package statesync

import (
	"encoding/json"
	"os"
)

const HashLen int = 32

// Snapshot is the header of a snapshot file representing the snapshot of the database at a certain height.
// It contains the height, format, chunk count, hash, size, and name of the snapshot.
// WARNING: This struct CAN NOT be changed without breaking functionality,
// since it is used for communication between nodes.
type Snapshot struct {
	Height       uint64          `json:"height"`
	Format       uint32          `json:"format"`
	ChunkHashes  [][HashLen]byte `json:"chunk_hashes"`
	ChunkCount   uint32          `json:"chunk_count"`
	SnapshotHash []byte          `json:"hash"`
	SnapshotSize uint64          `json:"size"`
}

// SaveAs saves the snapshot header to a file.
func (s *Snapshot) SaveAs(file string) error {
	bts, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, bts, 0644)
}

// LoadSnapshot loads the metadata associated with a db snapshot.
// It reads the snapshot header file and returns the snapshot metadata
func loadSnapshot(headerFile string) (*Snapshot, error) {
	bts, err := os.ReadFile(headerFile)
	if err != nil {
		return nil, err
	}
	var snapshot Snapshot
	if err := json.Unmarshal(bts, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *Snapshot) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Snapshot) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}
