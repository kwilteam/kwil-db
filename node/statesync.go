package node

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// TODO: set appropriate limits
	catalogSendTimeout = 15 * time.Second
	chunkSendTimeout   = 45 * time.Second
	chunkGetTimeout    = 45 * time.Second
	snapshotGetTimeout = 45 * time.Second

	snapshotCatalogNS    = "snapshot-catalog" // namespace on which snapshot catalogs are advertised
	discoverSnapshotsMsg = "discover_snapshots"
)

type StateSyncService struct {
	db       DB
	dbConfig config.DBConfig

	host       host.Host
	discoverer discovery.Discovery

	discoveryTimeout time.Duration

	snapshotStore SnapshotStore

	snapshotPool *snapshotPool // resets with every discovery

	trustedProviderAddrs []string
	trustedProviders     []*peer.AddrInfo

	snapshotDir string
	enable      bool // attempts to bootstrap the node with a snapshot

	log log.Logger
}

// snapshotKey is a snapshot key used for lookups.
type snapshotKey [sha256.Size]byte

type snapshot struct {
	Height      uint64     `json:"height"`
	Format      uint32     `json:"format"`
	Chunks      uint32     `json:"chunks"`
	Hash        []byte     `json:"hash"`
	Size        uint64     `json:"size"`
	ChunkHashes [][32]byte `json:"chunk_hashes"`
	// Metadata    []byte // TODO: do we need this?

	AppHash []byte `json:"app_hash"`
}

type snapshotPool struct {
	mtx       sync.Mutex // RWMutex?? no
	snapshots map[snapshotKey]*snapshot
	providers map[snapshotKey][]peer.AddrInfo // TODO: do we need this? should we request from all the providers instead?

	// Snapshot keys that have been blacklisted due to failed attempts to retrieve them or invalid data.
	blacklist map[snapshotKey]struct{}

	peers []peer.AddrInfo // do we need this?
}

// Key generates a snapshot key, used for lookups. It takes into account not only the height and
// format, but also the chunks, hash, and chunkHashes in case peers have generated snapshots in a
// non-deterministic manner. All fields must be equal for the snapshot to be considered the same.
func (s *snapshot) Key() snapshotKey {
	// Hash.Write() never returns an error.
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%v:%v:%v", s.Height, s.Format, s.Chunks)))
	hasher.Write(s.Hash)

	for _, chunkHash := range s.ChunkHashes {
		hasher.Write(chunkHash[:])
	}

	var key snapshotKey
	copy(key[:], hasher.Sum(nil))
	return key
}

func NewStateSyncService(ctx context.Context, h host.Host, discoverer discovery.Discovery, store SnapshotStore, log log.Logger, addrs []string, dir string, enable bool, db DB) (*StateSyncService, error) {
	ss := &StateSyncService{
		host:                 h,
		discoverer:           discoverer,
		discoveryTimeout:     15 * time.Second,
		snapshotStore:        store,
		log:                  log,
		trustedProviderAddrs: addrs,
		trustedProviders:     make([]*peer.AddrInfo, 0, len(addrs)),
		snapshotDir:          dir,
		snapshotPool: &snapshotPool{
			snapshots: make(map[snapshotKey]*snapshot),
			providers: make(map[snapshotKey][]peer.AddrInfo),
			blacklist: make(map[snapshotKey]struct{}),
		},
		enable: enable,
		db:     db,
	}

	// connect to trusted providers
	// TODO: can parallelize this
	for _, provider := range addrs {
		// connect to the provider
		i, err := connectPeer(ctx, provider, h)
		if err != nil {
			log.Warn("failed to connect to trusted provider", "provider", provider, "error", err)
		}

		ss.trustedProviders = append(ss.trustedProviders, i)
	}

	// provide stream handler for snapshot catalogs requests and chunk requests
	h.SetStreamHandler(ProtocolIDSnapshotCatalog, ss.snapshotCatalogRequestHandler)
	h.SetStreamHandler(ProtocolIDSnapshotChunk, ss.snapshotChunkRequestHandler)
	h.SetStreamHandler(ProtocolIDSnapshotMeta, ss.snapshotMetadataRequestHandler)

	return ss, nil
}

/*
Statesync service is responsible for all the below tasks:
1. Discover snapshot providers (peer discover)
2. Advertise snapshot catalogs
3. Retrieve snapshot catalogs and pick the best snapshot within a discovery timeperiod
4. Retrieve snapshot chunks and restore the snapshot
*/

func (s *StateSyncService) snapshotCatalogRequestHandler(stream network.Stream) {
	// read request
	// send snapshot catalogs
	defer stream.Close()

	stream.SetReadDeadline(time.Now().Add(time.Second))

	req := make([]byte, len(discoverSnapshotsMsg))
	n, err := stream.Read(req)
	if err != nil {
		s.log.Warn("failed to read discover snapshots request", "error", err)
		return
	}

	if n == 0 {
		// no request, hung up
		return
	}

	if string(req) != discoverSnapshotsMsg {
		s.log.Warn("invalid discover snapshots request")
		return
	}

	snapshots := s.snapshotStore.ListSnapshots()
	if snapshots == nil {
		// nothing to send
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}

	// send the snapshot catalogs
	catalogs := make([]*snapshot, 0, len(snapshots))
	for i, snap := range snapshots {
		catalogs[i] = &snapshot{
			Height:      snap.Height,
			Format:      snap.Format,
			Chunks:      snap.ChunkCount,
			Hash:        snap.SnapshotHash,
			ChunkHashes: make([][32]byte, snap.ChunkCount),
		}

		for j, chunk := range snap.ChunkHashes {
			copy(catalogs[i].ChunkHashes[j][:], chunk[:])
		}
	}

	encoder := json.NewEncoder(stream)
	stream.SetWriteDeadline(time.Now().Add(catalogSendTimeout))
	if err := encoder.Encode(catalogs); err != nil {
		s.log.Warn("failed to send snapshot catalogs", "error", err)
		return
	}

	s.log.Info("sent snapshot catalogs to remote peer", "peer", stream.Conn().RemotePeer(), "num_snapshots", len(catalogs))
}

func (s *StateSyncService) snapshotChunkRequestHandler(stream network.Stream) {
	// read request
	// send snapshot chunk
	defer stream.Close()

	stream.SetReadDeadline(time.Now().Add(chunkGetTimeout))
	var req snapshotChunkReq
	if _, err := req.ReadFrom(stream); err != nil {
		s.log.Warn("failed to read snapshot chunk request", "error", err)
		return
	}

	// read the snapshot chunk from the store
	chunk, err := s.snapshotStore.LoadSnapshotChunk(req.Height, req.Format, req.Index)
	if err != nil {
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}

	// send the snapshot chunk
	stream.SetWriteDeadline(time.Now().Add(chunkSendTimeout))
	stream.Write(chunk)

	s.log.Info("sent snapshot chunk to remote peer", "peer", stream.Conn().RemotePeer(), "height", req.Height, "index", req.Index)
}

func (s *StateSyncService) snapshotMetadataRequestHandler(stream network.Stream) {
	// read request
	// send snapshot chunk
	defer stream.Close()

	stream.SetReadDeadline(time.Now().Add(chunkGetTimeout))
	var req snapshotReq
	if _, err := req.ReadFrom(stream); err != nil {
		s.log.Warn("failed to read snapshot request", "error", err)
		return
	}

	// read the snapshot chunk from the store
	snap := s.snapshotStore.GetSnapshot(req.Height, req.Format)
	if snap == nil {
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}

	meta := &snapshot{
		Height:      snap.Height,
		Format:      snap.Format,
		Chunks:      snap.ChunkCount,
		Hash:        snap.SnapshotHash,
		ChunkHashes: make([][32]byte, snap.ChunkCount),
		Size:        snap.SnapshotSize,
	}
	for i, chunk := range snap.ChunkHashes {
		copy(meta.ChunkHashes[i][:], chunk[:])
	}

	// send the snapshot data
	encoder := json.NewEncoder(stream)

	stream.SetWriteDeadline(time.Now().Add(chunkSendTimeout))
	if err := encoder.Encode(meta); err != nil {
		s.log.Warn("failed to send snapshot metadata", "error", err)
		return
	}

	s.log.Info("sent snapshot chunk to remote peer", "peer", stream.Conn().RemotePeer(), "height", req.Height, "format", req.Format)
}

// verifySnapshot verifies the snapshot with the trusted provider
func (ss *StateSyncService) VerifySnapshot(ctx context.Context, snap *snapshot) bool {
	// verify the snapshot
	for _, provider := range ss.trustedProviders {
		// request the snapshot from the provider and verify the contents of the snapshot
		stream, err := ss.host.NewStream(ctx, provider.ID, ProtocolIDSnapshotMeta)
		if err != nil {
			ss.log.Warn("failed to request snapshot meta", "provider", provider.ID.String(), "error", err)
			continue
		}

		// request for the snapshot metadata
		req := snapshotReq{
			Height: snap.Height,
			Format: snap.Format,
		}
		reqBts, _ := req.MarshalBinary()
		stream.SetWriteDeadline(time.Now().Add(catalogSendTimeout))

		if _, err := stream.Write(reqBts); err != nil {
			ss.log.Warn("failed to send snapshot request", "provider", provider.ID.String(), "error", err)
			stream.Close()
			continue
		}

		stream.SetReadDeadline(time.Now().Add(snapshotGetTimeout))
		var meta snapshot
		if err := json.NewDecoder(stream).Decode(&meta); err != nil {
			ss.log.Warn("failed to decode snapshot metadata", "provider", provider.ID.String(), "error", err)
			stream.Close()
			continue
		}
		stream.Close()

		// verify the snapshot metadata
		if snap.Height != meta.Height || snap.Format != meta.Format || snap.Chunks != meta.Chunks {
			ss.log.Warnf("snapshot metadata mismatch: expected %v, got %v", snap, meta)
			continue
		}

		// snapshot hashes should match
		if !bytes.Equal(snap.Hash, meta.Hash) {
			ss.log.Warnf("snapshot metadata mismatch: expected %v, got %v", snap, meta)
			continue
		}

		// chunk hashes should match
		for i, chunkHash := range snap.ChunkHashes {
			if !bytes.Equal(chunkHash[:], meta.ChunkHashes[i][:]) {
				ss.log.Warnf("snapshot metadata mismatch: expected %v, got %v", snap, meta)
				continue
			}
		}

		ss.log.Info("verified snapshot with trusted provider", "provider", provider.ID.String())
		return true
	}
	return false
}

func (s *StateSyncService) blacklistSnapshot(snap *snapshot) {
	s.snapshotPool.mtx.Lock()
	defer s.snapshotPool.mtx.Unlock()

	key := snap.Key()
	s.snapshotPool.blacklist[key] = struct{}{}
	// delete the snapshot from the pool
	delete(s.snapshotPool.snapshots, key)
	delete(s.snapshotPool.providers, key)
}

func (s *StateSyncService) updatePeers(peers []peer.AddrInfo) {
	s.snapshotPool.mtx.Lock()
	defer s.snapshotPool.mtx.Unlock()

	s.snapshotPool.peers = peers
}

func (s *StateSyncService) getPeers() []peer.AddrInfo {
	s.snapshotPool.mtx.Lock()
	defer s.snapshotPool.mtx.Unlock()

	return s.snapshotPool.peers
}

func (s *StateSyncService) listSnapshots() []*snapshot {
	s.snapshotPool.mtx.Lock()
	defer s.snapshotPool.mtx.Unlock()

	snapshots := make([]*snapshot, 0, len(s.snapshotPool.snapshots))
	for _, snap := range s.snapshotPool.snapshots {
		snapshots = append(snapshots, snap)
	}
	return snapshots
}
