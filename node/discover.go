package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/node/peers"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// peerDiscoveryStreamHandler sends a list of peer addresses on the stream. This
// implements the receiving side of ProtocolIDDiscover.
func (n *Node) peerDiscoveryStreamHandler(s network.Stream) {
	defer s.Close()

	peers := n.pm.ConnectedPeers()

	s.SetWriteDeadline(time.Now().Add(4 * time.Second))
	if err := writePeers(s, peers); err != nil {
		n.log.Warn("failed to send peer list to peer", "error", err)
		return
	}

	n.log.Info("sent peer list to remote peer", "num_peers", len(peers),
		"to_peer", s.Conn().RemotePeer())
}

func writePeers(s io.WriteCloser, peers []peers.PeerInfo) error {
	encoder := json.NewEncoder(s)
	if err := encoder.Encode(peers); err != nil {
		return fmt.Errorf("failed to encode peers: %w", err)
	}
	return nil
}

func requestPeers(ctx context.Context, peerID peer.ID, host host.Host, log log.Logger) ([]peer.AddrInfo, error) {
	if peerID == host.ID() {
		return nil, nil
	}

	log.Info("Requesting peers", "from", peerID.String())

	stream, err := host.NewStream(ctx, peerID, ProtocolIDDiscover)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}

	stream.SetDeadline(deadline)

	return requestPeersProto(stream)
}

// requestPeersProto reads a list of peer addresses from the stream.
// This implements the initiating side of ProtocolIDDiscover.
func requestPeersProto(stream io.Reader) ([]peer.AddrInfo, error) {
	var peers []peer.AddrInfo
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&peers); err != nil {
		return nil, fmt.Errorf("failed to read and decode peers: %w", err)
	}
	return peers, nil
}
