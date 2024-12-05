package node

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/peers"
	"github.com/kwilteam/kwil-db/node/types"
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

type blockStore interface {
	GetByHeight(height int64) (types.Hash, *types.Block, types.Hash, error)
}

type statesyncConfig struct {
	StateSyncCfg *config.StateSyncConfig
	DBConfig     *config.DBConfig
	RcvdSnapsDir string

	DB            DB
	Host          host.Host
	Discoverer    discovery.Discovery
	SnapshotStore SnapshotStore
	BlockStore    blockStore
	Logger        log.Logger
}

type StateSyncService struct {
	// Config
	cfg              *config.StateSyncConfig
	dbConfig         *config.DBConfig
	snapshotDir      string
	trustedProviders []*peer.AddrInfo // trusted providers

	// DHT
	host       host.Host
	discoverer discovery.Discovery

	// Interfaces
	db            DB
	snapshotStore SnapshotStore
	blockStore    blockStore

	// statesync operation specific fields
	snapshotPool *snapshotPool // resets with every discovery

	// Logger
	log log.Logger
}

func NewStateSyncService(ctx context.Context, cfg *statesyncConfig) (*StateSyncService, error) {
	if cfg.StateSyncCfg.Enable && cfg.StateSyncCfg.TrustedProviders == nil {
		return nil, fmt.Errorf("at least one trusted provider is required for state sync")
	}

	ss := &StateSyncService{
		cfg:           cfg.StateSyncCfg,
		dbConfig:      cfg.DBConfig,
		snapshotDir:   cfg.RcvdSnapsDir,
		db:            cfg.DB,
		host:          cfg.Host,
		discoverer:    cfg.Discoverer,
		snapshotStore: cfg.SnapshotStore,
		log:           cfg.Logger,
		blockStore:    cfg.BlockStore,
		snapshotPool: &snapshotPool{
			snapshots: make(map[snapshotKey]*snapshotMetadata),
			providers: make(map[snapshotKey][]peer.AddrInfo),
			blacklist: make(map[snapshotKey]struct{}),
		},
	}

	// remove the existing snapshot directory
	if err := os.RemoveAll(ss.snapshotDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(ss.snapshotDir, 0755); err != nil {
		return nil, err
	}

	// provide stream handler for snapshot catalogs requests and chunk requests
	ss.host.SetStreamHandler(ProtocolIDSnapshotCatalog, ss.snapshotCatalogRequestHandler)
	ss.host.SetStreamHandler(ProtocolIDSnapshotChunk, ss.snapshotChunkRequestHandler)
	ss.host.SetStreamHandler(ProtocolIDSnapshotMeta, ss.snapshotMetadataRequestHandler)

	return ss, nil
}

func (s *StateSyncService) Bootstrap(ctx context.Context) error {
	providers, err := peers.ConvertPeersToMultiAddr(s.cfg.TrustedProviders)
	if err != nil {
		return err
	}

	for _, provider := range providers {
		// connect to the provider
		i, err := connectPeer(ctx, provider, s.host)
		if err != nil {
			s.log.Warn("failed to connect to trusted provider", "provider", provider, "error", err)
		}

		s.trustedProviders = append(s.trustedProviders, i)
	}
	return nil
}

// snapshotCatalogRequestHandler handles the incoming snapshot catalog requests.
// It sends the list of metadata of all the snapshots that are available with the node.
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
	catalogs := make([]*snapshotMetadata, len(snapshots))
	for i, snap := range snapshots {
		catalogs[i] = &snapshotMetadata{
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

// snapshotChunkRequestHandler handles the incoming snapshot chunk requests.
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

// snapshotMetadataRequestHandler handles the incoming snapshot metadata request and
// sends the snapshot metadata at the requested height.
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

	meta := &snapshotMetadata{
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

	// get the app hash from the db
	_, _, appHash, err := s.blockStore.GetByHeight(int64(snap.Height))
	if err != nil {
		s.log.Warn("failed to get app hash", "height", snap.Height, "error", err)
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}
	meta.AppHash = appHash[:]

	// send the snapshot data
	encoder := json.NewEncoder(stream)

	stream.SetWriteDeadline(time.Now().Add(chunkSendTimeout))
	if err := encoder.Encode(meta); err != nil {
		s.log.Warn("failed to send snapshot metadata", "error", err)
		return
	}

	s.log.Info("sent snapshot metadata to remote peer", "peer", stream.Conn().RemotePeer(), "height", req.Height, "format", req.Format, "appHash", appHash.String())
}

// verifySnapshot verifies the snapshot with the trusted provider and returns the app hash if the snapshot is valid.
func (ss *StateSyncService) VerifySnapshot(ctx context.Context, snap *snapshotMetadata) (bool, []byte) {
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
		var meta snapshotMetadata
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

		ss.log.Info("verified snapshot with trusted provider", "provider", provider.ID.String(), "snapshot", snap,
			"appHash", hex.EncodeToString(meta.AppHash))
		return true, meta.AppHash
	}
	return false, nil
}

type snapshotMetadata struct {
	Height      uint64     `json:"height"`
	Format      uint32     `json:"format"`
	Chunks      uint32     `json:"chunks"`
	Hash        []byte     `json:"hash"`
	Size        uint64     `json:"size"`
	ChunkHashes [][32]byte `json:"chunk_hashes"`

	AppHash []byte `json:"app_hash"`
}

func (sm *snapshotMetadata) String() string {
	return fmt.Sprintf("SnapshotMetadata{Height: %d, Format: %d, Chunks: %d, Hash: %x, Size: %d, AppHash: %x}", sm.Height, sm.Format, sm.Chunks, sm.Hash, sm.Size, sm.AppHash)
}

// snapshotKey is a snapshot key used for lookups.
type snapshotKey [sha256.Size]byte

// Key generates a snapshot key, used for lookups. It takes into account not only the height and
// format, but also the chunks, snapshot hash and chunk hashes in case peers have generated snapshots in a
// non-deterministic manner. All fields must be equal for the snapshot to be considered the same.
func (s *snapshotMetadata) Key() snapshotKey {
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

// snapshotPool keeps track of snapshots that have been discovered from the snapshot providers.
// It also keeps track of the providers that have advertised the snapshots and the blacklisted snapshots.
// Each snapshot is identified by a snapshot key which is generated from the snapshot metadata.
type snapshotPool struct {
	mtx       sync.Mutex // RWMutex?? no
	snapshots map[snapshotKey]*snapshotMetadata
	providers map[snapshotKey][]peer.AddrInfo // TODO: do we need this? should we request from all the providers instead?

	// Snapshot keys that have been blacklisted due to failed attempts to retrieve them or invalid data.
	blacklist map[snapshotKey]struct{}

	peers []peer.AddrInfo // do we need this?
}

func (sp *snapshotPool) blacklistSnapshot(snap *snapshotMetadata) {
	sp.mtx.Lock()
	defer sp.mtx.Unlock()

	key := snap.Key()
	sp.blacklist[key] = struct{}{}
	// delete the snapshot from the pool
	delete(sp.snapshots, key)
	delete(sp.providers, key)
}

func (sp *snapshotPool) updatePeers(peers []peer.AddrInfo) {
	sp.mtx.Lock()
	defer sp.mtx.Unlock()

	sp.peers = peers
}

func (sp *snapshotPool) keyProviders(key snapshotKey) []peer.AddrInfo {
	sp.mtx.Lock()
	defer sp.mtx.Unlock()

	return sp.providers[key]
}

func (sp *snapshotPool) getPeers() []peer.AddrInfo {
	sp.mtx.Lock()
	defer sp.mtx.Unlock()

	return sp.peers
}

func (sp *snapshotPool) listSnapshots() []*snapshotMetadata {
	sp.mtx.Lock()
	defer sp.mtx.Unlock()

	snapshots := make([]*snapshotMetadata, 0, len(sp.snapshots))
	for _, snap := range sp.snapshots {
		snapshots = append(snapshots, snap)
	}
	return snapshots
}
