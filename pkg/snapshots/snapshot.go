package snapshots

type Snapshot struct {
	Height     uint64
	Format     uint32
	ChunkCount uint32
	Hash       []byte

	Metadata SnapshotMetadata
}

type SnapshotMetadata struct {
	ChunkHashes map[uint32][]byte
	FileInfo    map[string]SnapshotFileInfo
}

type SnapshotFileInfo struct {
	Size     uint64
	Hash     []byte
	BeginIdx uint32
	EndIdx   uint32
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
