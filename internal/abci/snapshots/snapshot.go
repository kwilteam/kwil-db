package snapshots

type Snapshot struct {
	Height     uint64 `json:"height"`
	Format     uint32 `json:"format"`
	ChunkCount uint32 `json:"chunk_count"`
	Hash       []byte `json:"hash"`

	Metadata SnapshotMetadata
}

type SnapshotMetadata struct {
	ChunkHashes map[uint32][]byte           `json:"chunk_hashes"`
	FileInfo    map[string]SnapshotFileInfo `json:"file_info"`
}

type SnapshotFileInfo struct {
	Size     uint64 `json:"size"`
	Hash     []byte `json:"hash"`
	BeginIdx uint32 `json:"begin_idx"`
	EndIdx   uint32 `json:"end_idx"`
}

func (s *Snapshot) GetChunkHash(chunkID uint32) []byte {
	return s.Metadata.ChunkHashes[chunkID]
}

func (s *Snapshot) SnapshotMetadata() SnapshotMetadata {
	return s.Metadata
}

func (s *Snapshot) SnapshotFileInfo(file string) SnapshotFileInfo {
	return s.Metadata.FileInfo[file]
}
