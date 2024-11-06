package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"kwil/log"
	"kwil/node/types"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) peerDiscoveryStreamHandler(s network.Stream) {
	defer s.Close()

	s.SetReadDeadline(time.Now().Add(time.Second))

	buf := make([]byte, len(discoverPeersMsg))
	nr, err := s.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		n.log.Warn("failed to read peer discovery request", "error", err)
		return
	}
	if nr == 0 { // they hung up
		return
	}
	if string(buf) != discoverPeersMsg {
		n.log.Warn("invalid discover peers request")
		return
	}

	peers := n.pm.ConnectedPeers()

	s.SetWriteDeadline(time.Now().Add(4 * time.Second))
	if err := writePeers(s, peers); err != nil {
		n.log.Warn("failed to send peer list to peer", "error", err)
		return
	}

	n.log.Info("sent peer list to remote peer", "num_peers", len(peers),
		"to_peer", s.Conn().RemotePeer())
}

func writePeers(s io.WriteCloser, peers []types.PeerInfo) error {
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

	log.Info("Requesting peers from", "peer", peerID.String())

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

	if _, err := stream.Write([]byte(discoverPeersMsg)); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var peers []peer.AddrInfo
	if err := readPeersFromStream(stream, &peers); err != nil {
		return nil, fmt.Errorf("failed to read peers from stream: %w", err)
	}
	return peers, nil
}

// readPeersFromStream reads a list of peer addresses from the stream.
func readPeersFromStream(s network.Stream, peers *[]peer.AddrInfo) error {
	decoder := json.NewDecoder(s)
	if err := decoder.Decode(peers); err != nil {
		return fmt.Errorf("failed to decode peers: %w", err)
	}
	return nil
}

/*
type PeerMetadata struct {
	Info peer.AddrInfo
	Type string // e.g., "leader", "validator", "sentry"
}

// filterPeersForNodeType filters peers based on the requesting node's type.
func filterPeersForNodeType(peers []PeerMetadata, requesterType string) []PeerMetadata {
	var filtered []PeerMetadata

	for _, p := range peers {
		// share "leader" peers only with "validators"
		if p.Type == "leader" && requesterType != "validator" {
			continue // don't share "leader" peer info with non-validator nodes
		}

		// Add other policies as needed, e.g., only sharing "sentry" nodes with "leader" nodes

		// If the peer passes the filtering conditions, add it to the list
		filtered = append(filtered, p)
	}

	return filtered
}
*/
