package node

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/log"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/snapshotter"
	"github.com/kwilteam/kwil-db/node/types"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	data = sha256.Sum256([]byte("snapshot"))

	snap1 = &snapshotMetadata{
		Height:      1,
		Format:      1,
		Chunks:      1,
		Hash:        data[:],
		Size:        100,
		ChunkHashes: [][32]byte{data},
	}

	snap2 = &snapshotMetadata{
		Height:      2,
		Format:      1,
		Chunks:      1,
		Hash:        []byte("snap2"),
		Size:        100,
		ChunkHashes: [][32]byte{data},
	}
)

func newTestStatesyncer(ctx context.Context, t *testing.T, mn mock.Mocknet, rootDir string, sCfg *config.StateSyncConfig) (host.Host, discovery.Discovery, *snapshotStore, *StateSyncService, crypto.PrivateKey, error) {
	pkBts, h := newTestHost(t, mn)
	pk, err := crypto.UnmarshalSecp256k1PrivateKey(pkBts)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	dht, err := makeDHT(ctx, h, nil, dht.ModeServer)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	t.Cleanup(func() { dht.Close() })
	discover := makeDiscovery(dht)

	os.MkdirAll(rootDir, os.ModePerm)

	bs := &mockBS{}
	st := newSnapshotStore(bs)
	cfg := &StatesyncConfig{
		StateSyncCfg: sCfg,
		RcvdSnapsDir: rootDir,

		// DB, DBConfig unused
		P2PService:    &P2PService{host: h, dht: dht, discovery: discover},
		SnapshotStore: st,
		BlockStore:    bs,
		Logger:        log.DiscardLogger,
	}

	ss, err := NewStateSyncService(ctx, cfg)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	h.SetStreamHandler(snapshotter.ProtocolIDSnapshotCatalog, st.snapshotCatalogRequestHandler)
	h.SetStreamHandler(snapshotter.ProtocolIDSnapshotChunk, st.snapshotChunkRequestHandler)
	h.SetStreamHandler(snapshotter.ProtocolIDSnapshotMeta, st.snapshotMetadataRequestHandler)

	return h, discover, st, ss, pk, nil
}

func testSSConfig(enable bool, providers []string) *config.StateSyncConfig {
	return &config.StateSyncConfig{
		Enable:           enable,
		TrustedProviders: providers,
		DiscoveryTimeout: ktypes.Duration(5 * time.Second),
		MaxRetries:       3,
	}
}

func TestStateSyncService(t *testing.T) {
	ctx := context.Background()
	mn := mock.New()
	tempDir := t.TempDir()

	// trusted snapshot provider and statesync catalog service provider
	h1, d1, st1, _, pk1, err := newTestStatesyncer(ctx, t, mn, filepath.Join(tempDir, "n1"), testSSConfig(false, nil))
	require.NoError(t, err, "Failed to create statesyncer 1")

	bootPeer := fmt.Sprintf("%s#%s@127.0.0.1:6600", hex.EncodeToString(pk1.Public().Bytes()), pk1.Type())
	// statesync catalog service provider
	_, d2, st2, _, _, err := newTestStatesyncer(ctx, t, mn, filepath.Join(tempDir, "n2"), testSSConfig(false, nil))
	require.NoError(t, err, "Failed to create statesyncer 2")

	// node attempting statesync
	addrs := maddrs(h1)
	h3, d3, _, ss3, _, err := newTestStatesyncer(ctx, t, mn, filepath.Join(tempDir, "n3"), testSSConfig(true, []string{bootPeer}))
	require.NoError(t, err, "Failed to create statesyncer 3")

	// Link and connect the hosts
	err = mn.LinkAll()
	require.NoError(t, err, "Failed to link hosts")

	err = mn.ConnectAllButSelf()
	require.NoError(t, err, "Failed to connect hosts")

	// d1 and d2 advertise the snapshot catalog service
	advertise(ctx, snapshotCatalogNS, d1)
	advertise(ctx, snapshotCatalogNS, d2)

	time.Sleep(2 * time.Second)

	// bootstrap the ss3 with the trusted providers
	for _, addr := range addrs {
		i, err := connectPeer(ctx, addr, h3)
		require.NoError(t, err)
		ss3.trustedProviders = append(ss3.trustedProviders, i)
	}

	// h2 has a snapshot
	st2.addSnapshot(snap1)

	// Discover the snapshot catalog services
	peers, err := discoverProviders(ctx, snapshotCatalogNS, d1)
	require.NoError(t, err)
	peers = filterLocalPeer(peers, h1.ID())
	require.Len(t, peers, 1)

	peers, err = discoverProviders(ctx, snapshotCatalogNS, d3)
	require.NoError(t, err)
	peers = filterLocalPeer(peers, h3.ID())
	require.Len(t, peers, 2)

	// Request the snapshot catalogs
	for _, p := range peers {
		err = ss3.requestSnapshotCatalogs(ctx, p)
		require.NoError(t, err)
	}

	// should receive the snapshot catalog: snap1 from h2
	snaps := ss3.snapshotPool.listSnapshots()
	require.Len(t, snaps, 1)

	// best snapshot should be snap1
	bestSnap, err := ss3.bestSnapshot()
	require.NoError(t, err)
	assert.Equal(t, snap1.Height, bestSnap.Height)
	assert.Equal(t, snap1.Hash, bestSnap.Hash)

	// Validate the snapshot should fail as the trusted provider does not have the snapshot
	valid, _ := ss3.VerifySnapshot(ctx, snap1)
	assert.False(t, valid)

	// add snap1 to the trusted provider
	st1.addSnapshot(snap1)

	valid, _ = ss3.VerifySnapshot(ctx, snap1)
	assert.True(t, valid)

	// add snap2 to the trusted provider
	st1.addSnapshot(snap2)

	// best snapshot should be snap2
	for _, p := range peers {
		err = ss3.requestSnapshotCatalogs(ctx, p)
		require.NoError(t, err)
	}

	bestSnap, err = ss3.bestSnapshot()
	require.NoError(t, err)
	assert.Equal(t, snap2.Height, bestSnap.Height)

	valid, _ = ss3.VerifySnapshot(ctx, bestSnap)
	assert.True(t, valid)
}

type mockBS struct {
}

func (m *mockBS) GetByHeight(height int64) (types.Hash, *ktypes.Block, *types.CommitInfo, error) {
	return types.Hash{}, nil, &types.CommitInfo{AppHash: types.Hash{}}, nil
}

func (m *mockBS) Store(*ktypes.Block, *types.CommitInfo) error {
	return nil
}

func (m *mockBS) Best() (int64, types.Hash, types.Hash, time.Time) {
	return 0, types.Hash{}, types.Hash{}, time.Time{}
}

type snapshotStore struct {
	snapshots map[uint64]*snapshotMetadata
	bs        blockStore
}

func newSnapshotStore(bs blockStore) *snapshotStore {
	return &snapshotStore{
		snapshots: make(map[uint64]*snapshotMetadata),
		bs:        bs,
	}
}

func (s *snapshotStore) addSnapshot(snapshot *snapshotMetadata) {
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

func (s *snapshotStore) Enabled() bool {
	return true
}

func (s *snapshotStore) IsSnapshotDue(height uint64) bool {
	return false
}

func (s *snapshotStore) CreateSnapshot(ctx context.Context, height uint64, snapshotID string, schemas, excludedTables []string, excludeTableData []string) error {
	return nil
}

func (s *snapshotStore) snapshotCatalogRequestHandler(stream network.Stream) {
	defer stream.Close()
	stream.SetReadDeadline(time.Now().Add(time.Second))

	req := make([]byte, len(snapshotter.DiscoverSnapshotsMsg))
	n, err := stream.Read(req)
	if err != nil {
		return
	}

	if n == 0 { // no request, hung up
		return
	}

	snapshots := s.ListSnapshots()
	if snapshots == nil { // nothing to send
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}

	// send the snapshot catalogs
	catalogs := make([]*snapshotter.SnapshotMetadata, len(snapshots))
	for i, snap := range snapshots {
		catalogs[i] = snapshotToMetadata(snap)
	}

	encoder := json.NewEncoder(stream)
	stream.SetWriteDeadline(time.Now().Add(catalogSendTimeout))
	if err := encoder.Encode(catalogs); err != nil {
		return
	}
}

func (s *snapshotStore) snapshotChunkRequestHandler(stream network.Stream) {
	defer stream.Close()
	stream.SetReadDeadline(time.Now().Add(chunkGetTimeout))
	var req snapshotter.SnapshotChunkReq
	if _, err := req.ReadFrom(stream); err != nil {
		return
	}
	chunk, err := s.LoadSnapshotChunk(req.Height, req.Format, req.Index)
	if err != nil {
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}
	stream.SetWriteDeadline(time.Now().Add(chunkSendTimeout))
	stream.Write(chunk)
}

func (s *snapshotStore) snapshotMetadataRequestHandler(stream network.Stream) {
	defer stream.Close()
	stream.SetReadDeadline(time.Now().Add(chunkGetTimeout))
	var req snapshotter.SnapshotReq
	if _, err := req.ReadFrom(stream); err != nil {
		return
	}
	snap := s.GetSnapshot(req.Height, req.Format)
	if snap == nil {
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}

	meta := snapshotToMetadata(snap)

	// get the app hash from the db
	_, _, ci, err := s.bs.GetByHeight(int64(snap.Height))
	if err != nil || ci == nil {
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}
	meta.AppHash = ci.AppHash[:]

	// send the snapshot data
	encoder := json.NewEncoder(stream)

	stream.SetWriteDeadline(time.Now().Add(chunkSendTimeout))
	if err := encoder.Encode(meta); err != nil {
		return
	}
}

func snapshotToMetadata(s *snapshotter.Snapshot) *snapshotter.SnapshotMetadata {
	meta := &snapshotter.SnapshotMetadata{
		Height:      s.Height,
		Format:      s.Format,
		Chunks:      s.ChunkCount,
		Hash:        s.SnapshotHash,
		Size:        s.SnapshotSize,
		ChunkHashes: make([][32]byte, s.ChunkCount),
	}

	for i, chunk := range s.ChunkHashes {
		copy(meta.ChunkHashes[i][:], chunk[:])
	}

	return meta
}
