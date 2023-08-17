package snapshots

import (
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/snapshots"
)

type SnapshotStoreOpts func(*SnapshotStore)
type SnapshotStore struct {
	enabled         bool
	recurringHeight uint64
	maxSnapshots    uint64
	snapshotDir     string
	databaseDir     string
	chunkSize       uint64

	numSnapshots uint64
	log          log.Logger
	snapshotter  Snapshotter
}

func NewSnapshotStore(databaseDir string, snapshotDir string, height uint64, maxSnapshots uint64, opts ...SnapshotStoreOpts) *SnapshotStore {
	snapshotter := snapshots.NewSnapshotter(snapshotDir, databaseDir, 16*1024*1024)
	ss := &SnapshotStore{
		enabled:      true,
		numSnapshots: 0,
		chunkSize:    16 * 1024 * 1024,
		maxSnapshots: maxSnapshots,
		snapshotDir:  snapshotDir,
		databaseDir:  databaseDir,
		snapshotter:  snapshotter,
	}

	for _, opt := range opts {
		opt(ss)
	}
	return ss
}

func WithChunkSize(chunkSize uint64) SnapshotStoreOpts {
	return func(s *SnapshotStore) {
		s.chunkSize = chunkSize
	}
}

func WithLogger(logger log.Logger) SnapshotStoreOpts {
	return func(s *SnapshotStore) {
		s.log = logger
	}
}

func (s *SnapshotStore) IsSnapshotDue(height uint64) bool {
	return (height % s.recurringHeight) == 0
}

// Snapshot store Operations

// CreateSnapshot creates a snapshot at the given height & deletes the oldest snapshot if the max limit on snapshots has been reached
func (s *SnapshotStore) CreateSnapshot(height uint64) error {
	if !s.enabled {
		return nil
	}

	if s.snapshotter == nil {
		s.snapshotter = snapshots.NewSnapshotter(s.snapshotDir, s.databaseDir, s.chunkSize)
	}

	// Initialize snapshot session
	err := s.snapshotter.StartSnapshotSession(height)
	if err != nil {
		return err
	}

	// Create snapshot
	_ = s.snapshotter.CreateSnapshot()

	// Close snapshot session
	err = s.snapshotter.EndSnapshotSession()
	if err != nil {
		return err
	}

	s.numSnapshots++
	if s.numSnapshots > s.maxSnapshots {
		err = s.deleteOldestSnapshot()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SnapshotStore) ListSnapshots() ([]snapshots.Snapshot, error) {
	return s.snapshotter.ListSnapshots()
}

func (s *SnapshotStore) NumSnapshots() uint64 {
	return s.numSnapshots
}

func (s *SnapshotStore) LoadSnapshotChunk(height uint64, format uint32, chunkID uint32) []byte {
	chunk, err := s.snapshotter.LoadSnapshotChunk(height, format, chunkID)
	if err != nil {
		return nil
	}
	return chunk
}

func (s *SnapshotStore) deleteOldestSnapshot() error {
	err := s.snapshotter.DeleteOldestSnapshot()
	if err != nil {
		return err
	}
	s.numSnapshots--
	return nil
}
