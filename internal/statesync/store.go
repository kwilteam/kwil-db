package statesync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kwilteam/kwil-db/core/log"
)

/*
	SnapshotStore is responsible for creating, managing and serving the snapshots used for statesync process.
	The snapshots stored in the snapshot store cannot and should not be used for network migrations. As they
	don't always provide the latest state of the network and contains the state which shouldn't be available
	at the genesis of a new network.

	SnapshotStore Layout on disk:
	SnapshotsDir:
		snapshot-<height1>:
			snapshot-format-0
				header.json
				chunks:
					chunk-0.sql.gz
					chunk-1.sql.gz
					...
					chunk-n.sql.gz

		snapshot-<height2>:
			snapshot-format-0
				header.json
				chunks:
					chunk-0.sql.gz
					...
					chunk-n.sql.gz

	Currently we only support snapshots of format plain sql dump compressed with gzip.
	This can be extended to support other formats in the future.
*/

type DBConfig struct {
	DBUser string
	DBPass string
	DBHost string
	DBPort string
	DBName string
}

type SnapshotConfig struct {
	// Snapshot store configuration
	SnapshotDir     string
	MaxSnapshots    int
	RecurringHeight uint64
}

type SnapshotStore struct {
	// Snapshot Config
	cfg *SnapshotConfig

	// Snapshot Store
	snapshots       map[uint64]*Snapshot // Map of snapshot height to snapshot header
	snapshotHeights []uint64             // List of snapshot heights
	snapshotsMtx    sync.RWMutex         // Protects access to snapshots and snapshotHeights

	// Snapshotter
	snapshotter DBSnapshotter

	// Logger
	log log.Logger
}

type DBSnapshotter interface {
	CreateSnapshot(ctx context.Context, height uint64, snapshotID string) (*Snapshot, error)
}

func NewSnapshotStore(cfg *SnapshotConfig, dbCfg *DBConfig, logger log.Logger) (*SnapshotStore, error) {
	snapshotter := NewSnapshotter(dbCfg, cfg.SnapshotDir, logger)
	ss := &SnapshotStore{
		cfg:         cfg,
		snapshots:   make(map[uint64]*Snapshot),
		snapshotter: snapshotter,
		log:         logger,
	}

	err := ss.loadSnapshots()
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshots from disk: %w", err)
	}

	return ss, nil
}

// IsSnapshotDue checks if a snapshot is due at the given height.
func (s *SnapshotStore) IsSnapshotDue(height uint64) bool {
	return (height % s.cfg.RecurringHeight) == 0
}

// List snapshots lists all the registered snapshots in the snapshot store.
func (s *SnapshotStore) ListSnapshots() []*Snapshot {
	s.snapshotsMtx.RLock()
	defer s.snapshotsMtx.RUnlock()

	snaps := make([]*Snapshot, 0, len(s.snapshots))
	for _, snap := range s.snapshots {
		snaps = append(snaps, snap)
	}

	return snaps
}

// CreateSnapshot creates a new snapshot at the given height and snapshot ID.
// SnapshotStore ensures that the number of snapshots does not exceed the maximum configured snapshots.
// If exceeds, it deletes the oldest snapshot.
func (s *SnapshotStore) CreateSnapshot(ctx context.Context, height uint64, snapshotID string) error {
	// Create a snapshot of the database at the given height
	snapshot, err := s.snapshotter.CreateSnapshot(ctx, height, snapshotID)
	if err != nil {
		os.RemoveAll(snapshotHeightDir(s.cfg.SnapshotDir, height))
		return fmt.Errorf("failed to create snapshot at height %d: %w", height, err)
	}

	// Register the snapshot
	err = s.RegisterSnapshot(snapshot)
	if err != nil {
		os.RemoveAll(snapshotHeightDir(s.cfg.SnapshotDir, height))
		return fmt.Errorf("failed to register snapshot at height %d: %w", height, err)
	}

	return nil
}

// RegisterSnapshot registers the existing snapshot in the snapshot store.
// It ensures that the number of snapshots does not exceed the maximum configured snapshots.
// If exceeds, it deletes the oldest snapshot.
func (s *SnapshotStore) RegisterSnapshot(snapshot *Snapshot) error {
	s.snapshotsMtx.Lock()
	defer s.snapshotsMtx.Unlock()

	if snapshot == nil {
		return nil // no snapshot to register
	}

	if _, ok := s.snapshots[snapshot.Height]; ok { // snapshot already exists at the given height
		return nil
	}

	// Register the snapshot
	s.snapshots[snapshot.Height] = snapshot
	s.snapshotHeights = append(s.snapshotHeights, snapshot.Height)

	// Sort the snapshot heights in ascending order
	slices.Sort(s.snapshotHeights)

	// Check if the number of snapshots exceeds the maximum number of snapshots
	for len(s.snapshotHeights) > s.cfg.MaxSnapshots {
		// Delete the oldest snapshot
		s.deleteOldestSnapshot()
	}
	return nil
}

// DeleteOldestSnapshot deletes the oldest snapshot.
// Deletes the internal and fs snapshot files and references corresponding to the oldest snapshot.
func (s *SnapshotStore) deleteOldestSnapshot() error {
	if len(s.snapshotHeights) == 0 {
		return nil
	}

	oldHeight := s.snapshotHeights[0]
	snapshot := s.snapshots[oldHeight]
	snapshotDir := snapshotHeightDir(s.cfg.SnapshotDir, snapshot.Height)

	os.RemoveAll(snapshotDir) // Delete the oldest snapshot directory

	delete(s.snapshots, oldHeight)            // delete the snapshot reference
	s.snapshotHeights = s.snapshotHeights[1:] // remove the oldest snapshot height
	return nil
}

// LoadSnapshotChunk loads a snapshot chunk at the given height and chunk index of given format.
// It returns the snapshot chunk as a byte slice of max size 16MB.
// errors if the chunk of chunkIndex corresponding to snapshot at given height and format does not exist.
func (s *SnapshotStore) LoadSnapshotChunk(height uint64, format uint32, chunkIdx uint32) ([]byte, error) {
	s.snapshotsMtx.RLock()
	defer s.snapshotsMtx.RUnlock()

	// Check if snapshot format is supported
	if format != DefaultSnapshotFormat {
		return nil, fmt.Errorf("unsupported snapshot format %d", format)
	}

	// Check if snapshot exists
	snapshot, ok := s.snapshots[height]
	if !ok {
		return nil, fmt.Errorf("snapshot at height %d does not exist", height)
	}

	// Check if chunk exists
	if chunkIdx >= snapshot.ChunkCount {
		return nil, fmt.Errorf("chunk %d does not exist in snapshot at height %d", chunkIdx, height)
	}

	chunkFile := snapshotChunkFile(s.cfg.SnapshotDir, height, format, chunkIdx)
	if _, err := os.Open(chunkFile); err != nil {
		return nil, fmt.Errorf("chunk %d does not exist in snapshot at height %d", chunkIdx, height)
	}

	// Read the chunk file
	bts, err := os.ReadFile(chunkFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk %d at height %d: %w", chunkIdx, height, err)
	}

	return bts, nil
}

// LoadSnapshotHeader loads the snapshots from the snapshotDir
// and registers them in the snapshot store.
func (s *SnapshotStore) loadSnapshots() error {
	// Scan the snapshot directory and load all the snapshots
	files, err := os.ReadDir(s.cfg.SnapshotDir)
	if err != nil {
		return nil
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		fileName := file.Name() // format: block-<height>
		names := strings.Split(fileName, "-")
		if len(names) != 2 {
			s.log.Debug("invalid snapshot directory name, ignoring the snapshot", log.String("dir", fileName))
			continue
		}
		height := names[1]
		heightInt, err := strconv.ParseUint(height, 10, 64)
		if err != nil {
			s.log.Debug("invalid snapshot height, ignoring the snapshot", log.String("height", height))
		}

		// Load snapshot header
		headerFile := snapshotHeaderFile(s.cfg.SnapshotDir, heightInt, DefaultSnapshotFormat)
		header, err := loadSnapshot(headerFile)
		if err != nil {
			s.log.Debug("Invalid snapshot header file, ignoring the snapshot", log.String("height", height), log.String("Error", err.Error()))
		}

		// Ensure that the chunk files exist
		for i := uint32(0); i < header.ChunkCount; i++ {
			chunkFile := snapshotChunkFile(s.cfg.SnapshotDir, heightInt, DefaultSnapshotFormat, i)
			if _, err := os.Open(chunkFile); err != nil { // chunk file doesn't exist
				s.log.Debug("Invalid snapshot chunk file, ignoring the snapshot", log.String("chunk-file", chunkFile))
			}
		}

		s.snapshots[heightInt] = header
		s.snapshotHeights = append(s.snapshotHeights, heightInt)
	}

	sort.Slice(s.snapshotHeights, func(i, j int) bool {
		return s.snapshotHeights[i] < s.snapshotHeights[j]
	})

	// Check if the number of snapshots exceeds the maximum number of snapshots
	for len(s.snapshotHeights) > s.cfg.MaxSnapshots {
		// Delete the oldest snapshot
		if err := s.deleteOldestSnapshot(); err != nil {
			return fmt.Errorf("failed to delete oldest snapshot: %w", err)
		}
	}

	return nil
}

// utility functions
func snapshotHeightDir(snapshotDir string, height uint64) string {
	return filepath.Join(snapshotDir, fmt.Sprintf("block-%d", height))
}

func snapshotFormatDir(snapshotDir string, height uint64, format uint32) string {
	return filepath.Join(snapshotHeightDir(snapshotDir, height), fmt.Sprintf("format-%d", format))
}

func snapshotChunkDir(snapshotDir string, height uint64, format uint32) string {
	return filepath.Join(snapshotFormatDir(snapshotDir, height, format), "chunks")
}

func snapshotChunkFile(snapshotDir string, height uint64, format uint32, chunkIdx uint32) string {
	return filepath.Join(snapshotChunkDir(snapshotDir, height, format), fmt.Sprintf("chunk-%d.sql.gz", chunkIdx))
}

func snapshotHeaderFile(snapshotDir string, height uint64, format uint32) string {
	return filepath.Join(snapshotFormatDir(snapshotDir, height, format), "header.json")
}
