package snapshotter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

const (
	DiscoverSnapshotsMsg = "discover_snapshots"
	reqRWTimeout         = 15 * time.Second
	catalogSendTimeout   = 15 * time.Second
	chunkSendTimeout     = 45 * time.Second
	chunkGetTimeout      = 45 * time.Second
	snapshotGetTimeout   = 45 * time.Second

	ProtocolIDSnapshotCatalog protocol.ID = "/kwil/snapcat/1.0.0"
	ProtocolIDSnapshotChunk   protocol.ID = "/kwil/snapchunk/1.0.0"
	ProtocolIDSnapshotMeta    protocol.ID = "/kwil/snapmeta/1.0.0"

	SnapshotCatalogNS = "snapshot-catalog" // namespace on which snapshot catalogs are advertised
)

var noData = []byte{0}

// RegisterSnapshotStreamHandlers registers the snapshot stream handlers if snapshotting is enabled.
func (s *SnapshotStore) RegisterSnapshotStreamHandlers(ctx context.Context, host host.Host, discovery discovery.Discovery) {
	if s == nil || s.cfg == nil || !s.cfg.Enable {
		// return if snapshotting is disabled
		return
	}

	// Register snapshot stream handlers
	host.SetStreamHandler(ProtocolIDSnapshotCatalog, s.snapshotCatalogRequestHandler)
	host.SetStreamHandler(ProtocolIDSnapshotChunk, s.snapshotChunkRequestHandler)
	host.SetStreamHandler(ProtocolIDSnapshotMeta, s.snapshotMetadataRequestHandler)

	// Advertise the snapshotcatalog service if snapshots are enabled
	// umm, but gotcha, if a node has previous snapshots but snapshots are disabled, these snapshots will be unusable.
	util.Advertise(ctx, discovery, SnapshotCatalogNS)
}

// SnapshotCatalogRequestHandler handles the incoming snapshot catalog requests.
// It sends the list of metadata of all the snapshots that are available with the node.
func (s *SnapshotStore) snapshotCatalogRequestHandler(stream network.Stream) {
	// read request
	// send snapshot catalogs
	defer stream.Close()

	stream.SetReadDeadline(time.Now().Add(time.Second))

	req := make([]byte, len(DiscoverSnapshotsMsg))
	n, err := stream.Read(req)
	if err != nil {
		s.log.Warn("failed to read discover snapshots request", "error", err)
		return
	}

	if n == 0 {
		// no request, hung up
		return
	}

	if string(req) != DiscoverSnapshotsMsg {
		s.log.Warn("invalid discover snapshots request")
		return
	}

	snapshots := s.ListSnapshots()
	if snapshots == nil {
		// nothing to send
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}

	// send the snapshot catalogs
	catalogs := make([]*SnapshotMetadata, len(snapshots))
	for i, snap := range snapshots {
		catalogs[i] = &SnapshotMetadata{
			Height:      snap.Height,
			Format:      snap.Format,
			Chunks:      snap.ChunkCount,
			Hash:        snap.SnapshotHash,
			Size:        snap.SnapshotSize,
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

// SnapshotChunkRequestHandler handles the incoming snapshot chunk requests.
func (s *SnapshotStore) snapshotChunkRequestHandler(stream network.Stream) {
	// read request
	// send snapshot chunk
	defer stream.Close()

	stream.SetReadDeadline(time.Now().Add(chunkGetTimeout))
	var req SnapshotChunkReq
	if _, err := req.ReadFrom(stream); err != nil {
		s.log.Warn("failed to read snapshot chunk request", "error", err)
		return
	}

	// read the snapshot chunk from the store
	chunk, err := s.LoadSnapshotChunk(req.Height, req.Format, req.Index)
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

// SnapshotMetadataRequestHandler handles the incoming snapshot metadata request and
// sends the snapshot metadata at the requested height.
func (s *SnapshotStore) snapshotMetadataRequestHandler(stream network.Stream) {
	// read request
	// send snapshot chunk
	defer stream.Close()

	stream.SetReadDeadline(time.Now().Add(chunkGetTimeout))
	var req SnapshotReq
	if _, err := req.ReadFrom(stream); err != nil {
		s.log.Warn("failed to read snapshot request", "error", err)
		return
	}

	// read the snapshot chunk from the store
	snap := s.GetSnapshot(req.Height, req.Format)
	if snap == nil {
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}

	meta := &SnapshotMetadata{
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
	_, _, ci, err := s.blockStore.GetByHeight(int64(snap.Height))
	if err != nil || ci == nil {
		s.log.Warn("failed to get app hash", "height", snap.Height, "error", err)
		stream.SetWriteDeadline(time.Now().Add(reqRWTimeout))
		stream.Write(noData)
		return
	}
	meta.AppHash = ci.AppHash[:]

	// send the snapshot data
	encoder := json.NewEncoder(stream)

	stream.SetWriteDeadline(time.Now().Add(chunkSendTimeout))
	if err := encoder.Encode(meta); err != nil {
		s.log.Warn("failed to send snapshot metadata", "error", err)
		return
	}

	s.log.Info("sent snapshot metadata to remote peer", "peer", stream.Conn().RemotePeer(), "height", req.Height, "format", req.Format, "appHash", ci.AppHash.String())
}
