package node

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
)

var (
	data = sha256.Sum256([]byte("snapshot"))

	snap1 = &snapshot{
		Height:      1,
		Format:      1,
		Chunks:      1,
		Hash:        data[:],
		Size:        100,
		ChunkHashes: [][32]byte{data},
	}

	snap2 = &snapshot{
		Height:      2,
		Format:      1,
		Chunks:      1,
		Hash:        []byte("snap2"),
		Size:        100,
		ChunkHashes: [][32]byte{data},
	}

	invalidSnap1 = &snapshot{
		Height:      1,
		Format:      1,
		Chunks:      1,
		Hash:        []byte("snap1-invalid"),
		Size:        100,
		ChunkHashes: [][32]byte{data},
	}
)

type snapshotStore struct {
	snapshots map[uint64]*snapshot
}

func NewSnapshotStore() *snapshotStore {
	return &snapshotStore{
		snapshots: make(map[uint64]*snapshot),
	}
}

func (s *snapshotStore) addSnapshot(snapshot *snapshot) {
	s.snapshots[snapshot.Height] = snapshot
}

func (s *snapshotStore) ListSnapshots() []*snapshotter.Snapshot {
	snapshots := make([]*snapshotter.Snapshot, 0, len(s.snapshots))
	for _, snapshot := range s.snapshots {
		snap := &snapshotter.Snapshot{
			Height:       snapshot.Height,
			Format:       snapshot.Format,
			ChunkCount:   snapshot.Chunks,
			SnapshotSize: snapshot.Size,
			SnapshotHash: snapshot.Hash,
			ChunkHashes:  make([][32]byte, len(snapshot.ChunkHashes)),
		}

		for j, hash := range snapshot.ChunkHashes {
			copy(snap.ChunkHashes[j][:], hash[:])
		}

		snapshots = append(snapshots, snap)
	}
	return snapshots
}

func (s *snapshotStore) LoadSnapshotChunk(height uint64, format uint32, index uint32) ([]byte, error) {
	snapshot, ok := s.snapshots[height]
	if !ok {
		return nil, errors.New("snapshot not found")
	}

	if index >= snapshot.Chunks {
		return nil, errors.New("chunk not found")
	}

	return []byte("snapshot"), nil
}

func (s *snapshotStore) GetSnapshot(height uint64, format uint32) *snapshotter.Snapshot {
	snapshot, ok := s.snapshots[height]
	if !ok {
		return nil
	}

	return &snapshotter.Snapshot{
		Height:       snapshot.Height,
		Format:       snapshot.Format,
		ChunkCount:   snapshot.Chunks,
		SnapshotSize: snapshot.Size,
		SnapshotHash: snapshot.Hash,
		ChunkHashes:  snapshot.ChunkHashes,
	}
}

func TestValidateSnapshot(t *testing.T) {
	ctx := context.Background()
	mn := mock.New()
	// trusted snapshot provider and exposing the snapshot catalog service
	_, h1, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}
	_, h3, err := newTestHost(t, mn)
	if err != nil {
		t.Fatalf("Failed to add peer to mocknet: %v", err)
	}

	// Link and connect the hosts
	if err := mn.LinkAll(); err != nil {
		t.Fatalf("Failed to link hosts: %v", err)
	}
	if err := mn.ConnectAllButSelf(); err != nil {
		t.Fatalf("Failed to connect hosts: %v", err)
	}

	time.Sleep(time.Second)

	tempDir := t.TempDir()
	root1 := filepath.Join(tempDir, "snap1")
	root2 := filepath.Join(tempDir, "snap2")

	os.MkdirAll(root1, os.ModePerm)
	os.MkdirAll(root2, os.ModePerm)

	peerIDs1 := h1.Peerstore().Peers()
	assert.Len(t, peerIDs1, 2)
	var peers1 []peer.AddrInfo
	for _, p := range peerIDs1 {
		pi := h3.Peerstore().PeerInfo(p)
		peers1 = append(peers1, pi)
	}

	peerIDs3 := h3.Peerstore().Peers()
	assert.Len(t, peerIDs3, 2)
	var peers3 []peer.AddrInfo
	for _, p := range peerIDs3 {
		pi := h3.Peerstore().PeerInfo(p)
		peers3 = append(peers3, pi)
	}

	dht1, err := makeDHT(ctx, h1, peers1, dht.ModeServer)
	assert.NoError(t, err, "Failed to create DHT1")
	t.Cleanup(func() { dht1.Close() })
	discover1 := makeDiscovery(dht1)
	addrs := maddrs(h1)

	st1 := NewSnapshotStore()
	st1.addSnapshot(snap1)
	st1.addSnapshot(snap2)
	ss1 := &StateSyncService{
		host:             h1,
		discoverer:       discover1,
		discoveryTimeout: 15 * time.Second,
		snapshotStore:    st1,
		log:              log.DiscardLogger,
		snapshotPool: &snapshotPool{
			snapshots: make(map[snapshotKey]*snapshot),
			providers: make(map[snapshotKey][]peer.AddrInfo),
			blacklist: make(map[snapshotKey]struct{}),
		},
		snapshotDir: root1,
	}
	addStreamHandlers(h1, ss1)

	// new node trying to bootup using the snapshot

	st3 := NewSnapshotStore()
	dht3, err := makeDHT(ctx, h3, peers3, dht.ModeServer)
	assert.NoError(t, err, "Failed to create DHT3")
	t.Cleanup(func() { dht3.Close() })
	discover3 := makeDiscovery(dht3)
	ss3 := &StateSyncService{
		host:             h3,
		discoverer:       discover3,
		discoveryTimeout: 15 * time.Second,
		snapshotStore:    st3,
		log:              log.DiscardLogger,
		snapshotPool: &snapshotPool{
			snapshots: make(map[snapshotKey]*snapshot),
			providers: make(map[snapshotKey][]peer.AddrInfo),
			blacklist: make(map[snapshotKey]struct{}),
		},
		snapshotDir: root2,
	}
	addStreamHandlers(h3, ss3)

	// advertise the snapshot catalog service
	advertise(ctx, snapshotCatalogNS, discover1)
	advertise(ctx, snapshotCatalogNS, discover3)

	time.Sleep(5 * time.Second)

	// Validate the snapshot (no trusted providers to validate against)
	valid := ss3.VerifySnapshot(ctx, snap1)
	assert.False(t, valid)

	// add h1 as the trusted provider
	for _, addr := range addrs {
		i, err := connectPeer(ctx, addr, h3)
		assert.NoError(t, err)
		ss3.trustedProviders = append(ss3.trustedProviders, i)
	}

	// Validate the snapshot (trusted provider has the snapshot)
	valid = ss3.VerifySnapshot(ctx, snap1)
	assert.True(t, valid)

	valid = ss3.VerifySnapshot(ctx, invalidSnap1)
	assert.False(t, valid)

	// Discovery test
	peers3, err = discoverProviders(ctx, snapshotCatalogNS, discover3)
	assert.NoError(t, err)
	peers3 = filterLocalPeer(peers3, h3.ID())
	assert.Len(t, peers3, 1)

	peers1, err = discoverProviders(ctx, snapshotCatalogNS, discover1)
	assert.NoError(t, err)
	peers1 = filterLocalPeer(peers1, h1.ID())
	assert.Len(t, peers1, 1)

	err = ss3.requestSnapshotCatalogs(ctx, peers3[0])
	assert.NoError(t, err)
	// should receive the snapshot catalog: snap1, snap2
	snaps := ss3.listSnapshots()
	assert.Len(t, snaps, 2)

	bestSnap, err := ss3.bestSnapshot()
	assert.NoError(t, err)
	assert.Equal(t, snap2.Height, bestSnap.Height)

	valid = ss3.VerifySnapshot(ctx, bestSnap)
	assert.True(t, valid)

	// request the snapshot chunks
	err = ss3.chunkFetcher(ctx, bestSnap)
	assert.NoError(t, err)

	// ensure that chunks are downloaded
	for i := range bestSnap.Chunks {
		cfile := filepath.Join(root2, fmt.Sprintf("chunk-%d.sql.gz", i))
		f, err := os.Open(cfile)
		assert.NoError(t, err)
		assert.NotNil(t, f)
		f.Close()
	}
}

func addStreamHandlers(h host.Host, ss *StateSyncService) {
	h.SetStreamHandler(ProtocolIDSnapshotCatalog, ss.snapshotCatalogRequestHandler)
	h.SetStreamHandler(ProtocolIDSnapshotChunk, ss.snapshotChunkRequestHandler)
	h.SetStreamHandler(ProtocolIDSnapshotMeta, ss.snapshotMetadataRequestHandler)
}

func filterLocalPeer(peers []peer.AddrInfo, localID peer.ID) []peer.AddrInfo {
	var filteredPeers []peer.AddrInfo
	for _, p := range peers {
		if p.ID != localID {
			filteredPeers = append(filteredPeers, p)
		}
	}
	return filteredPeers
}
