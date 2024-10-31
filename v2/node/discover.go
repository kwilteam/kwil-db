package node

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"kwil/log"
	"kwil/node/types"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) peerDiscoveryStreamHandler(s network.Stream) {
	defer s.Close()

	sc := bufio.NewScanner(s)
	for sc.Scan() { // why am I doing this again? Probably just Read once...
		msg := sc.Text()
		if msg != discoverPeersMsg {
			continue
		}

		peers := n.pm.KnownPeers()
		// filteredPeers := filterPeersForNodeType(peers, nodeType)
		if err := sendPeersToStream(s, peers); err != nil {
			fmt.Println("failed to send peer list to peer", err)
			return
		}
	}

	// fmt.Println("done sending peers on stream", s.ID())
}

func sendPeersToStream(s network.Stream, peers []types.PeerInfo) error {
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

	requestMessage := discoverPeersMsg
	if _, err := stream.Write([]byte(requestMessage + "\n")); err != nil {
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
