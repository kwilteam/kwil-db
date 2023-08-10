package snapshots

import "github.com/kwilteam/kwil-db/pkg/snapshots"

type Snapshotter interface {
	StartSnapshotSession(height uint64) error
	EndSnapshotSession() error
	CreateSnapshot() error
	LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) ([]byte, error)
	DeleteOldestSnapshot() error
	ListSnapshots() ([]snapshots.Snapshot, error)
}
