package peers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	ProtocolIDDiscover protocol.ID = "/kwil/discovery/1.0.0" // the PEX protocol that all peers support, and is the only real protocol that crawlers support other than the dummy ProtocolIDCrawler
	ProtocolIDCrawler  protocol.ID = "/kwil/crawler/1.0.0"   // dummy to identify a crawler that should not stay connected
)

// DiscoveryStreamHandler sends a list of peer addresses on the stream. This
// implements the receiving side of ProtocolIDDiscover.
func (pm *PeerMan) DiscoveryStreamHandler(s network.Stream) {
	defer s.Close()

	pid := s.Conn().RemotePeer()
	if pm.seedMode { // in seed mode hang up after serving the peer list
		defer pm.h.Network().ClosePeer(pid)
	}

	peers := pm.ConnectedPeers()

	s.SetWriteDeadline(time.Now().Add(4 * time.Second))
	if err := writePeers(s, peers); err != nil {
		pm.log.Warn("failed to send peer list to peer", "error", err)
		return
	}

	pm.log.Debug("sent peer list to remote peer", "num_peers", len(peers),
		"to_peer", pid)
}

func writePeers(s io.WriteCloser, peers []PeerInfo) error {
	encoder := json.NewEncoder(s)
	if err := encoder.Encode(peers); err != nil {
		return fmt.Errorf("failed to encode peers: %w", err)
	}
	return nil
}

// RequestPeers initiates the ProtocolIDDiscover stream.
func (pm *PeerMan) RequestPeers(ctx context.Context, peerID peer.ID) ([]peer.AddrInfo, error) {
	if peerID == pm.h.ID() {
		return nil, nil
	}

	pm.log.Debug("Requesting peers", "from", peerID.String())

	stream, err := pm.h.NewStream(ctx, peerID, ProtocolIDDiscover)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", CompressDialError(err))
	}
	defer stream.Close()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}

	stream.SetDeadline(deadline)

	peers, err := recvPeersProto(stream)
	if err != nil {
		return nil, err
	}

	// Ensure that each received peer ID is an "identity" multihash that encodes
	// a supported public key.
	var okPeers []peer.AddrInfo
	for _, peer := range peers {
		pk, _ := pubKeyFromPeerID(peer.ID)
		if pk == nil {
			pm.log.Warnf("Invalid peer ID received from %v: %v", peerID, peer.ID)
		} else {
			okPeers = append(okPeers, peer)
		}
	}

	return okPeers, nil
}

// recvPeersProto reads a list of peer addresses from the stream.
// This implements the initiating side of ProtocolIDDiscover.
func recvPeersProto(stream io.Reader) ([]peer.AddrInfo, error) {
	var peers []peer.AddrInfo
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&peers); err != nil {
		return nil, fmt.Errorf("failed to read and decode peers: %w", err)
	}
	return peers, nil
}
