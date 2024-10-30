package node

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"kwil/log"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

var (
	discoverPeersMsg = "discover_peers"
)

type peerMan struct {
	log log.Logger
	h   host.Host
	// nodeType string // e.g., "leader", "validator", "sentry"
}

var _ discovery.Discoverer = (*peerMan)(nil) // FindPeers method

var _ network.StreamHandler = new(peerMan).discoveryStreamHandler

func (pm *peerMan) discoveryStreamHandler(s network.Stream) {
	defer s.Close()

	sc := bufio.NewScanner(s)
	for sc.Scan() { // why am I doing this again? Probably just Read once...
		msg := sc.Text()
		if msg != discoverPeersMsg {
			continue
		}

		peers := getKnownPeers(pm.h)
		// filteredPeers := filterPeersForNodeType(peers, nodeType)

		if err := sendPeersToStream(s, peers); err != nil {
			fmt.Println("failed to send peer list to peer", err)
			return
		}
	}

	// fmt.Println("done sending peers on stream", s.ID())
}

// func (pm *peerMan) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
// 	return 0, nil
// }

func (pm *peerMan) requestPeers(ctx context.Context, peerID peer.ID) ([]peer.AddrInfo, error) {
	if peerID == pm.h.ID() {
		return nil, nil
	}

	stream, err := pm.h.NewStream(ctx, peerID, ProtocolIDDiscover)
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

func (pm *peerMan) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	peerChan := make(chan peer.AddrInfo)

	peers := pm.h.Network().Peers()
	if len(peers) == 0 {
		close(peerChan)
		pm.log.Warn("no existing peers for peer discovery")
		return peerChan, nil
	}

	var wg sync.WaitGroup
	wg.Add(len(peers))
	for _, peerID := range peers {
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			peers, err := pm.requestPeers(ctx, peerID)
			if err != nil {
				fmt.Printf("Failed to get peers from %v: %v", peerID, err)
				return
			}

			for _, p := range peers {
				peerChan <- p
			}
		}()
	}

	go func() {
		wg.Wait()
		close(peerChan)
	}()

	return peerChan, nil
}

type AddrInfo struct {
	ID    peer.ID               `json:"id"`
	Addrs []multiaddr.Multiaddr `json:"addrs"`
}

type PeerInfo struct {
	AddrInfo
	Protos []protocol.ID `json:"protos"`
}

func getKnownPeers(h host.Host) []PeerInfo {
	var peers []PeerInfo
	for _, peerID := range h.Network().Peers() { // connected peers only
		addrs := h.Peerstore().Addrs(peerID)

		supportedProtos, err := h.Peerstore().GetProtocols(peerID)
		if err != nil {
			fmt.Printf("GetProtocols for %v: %v\n", peerID, err)
			continue
		}

		peers = append(peers, PeerInfo{
			AddrInfo: AddrInfo{
				ID:    peerID,
				Addrs: addrs,
			},
			Protos: supportedProtos,
		})

	}
	return peers
}

func sendPeersToStream(s network.Stream, peers []PeerInfo) error {
	encoder := json.NewEncoder(s)
	if err := encoder.Encode(peers); err != nil {
		return fmt.Errorf("failed to encode peers: %w", err)
	}
	return nil
}

// addPeerToPeerStore adds a discovered peer to the local peer store.
func addPeerToPeerStore(ps peerstore.Peerstore, p peer.AddrInfo) {
	// Only add the peer if it's not already in the peer store.
	// h.Peerstore().Peers()
	addrs := ps.Addrs(p.ID)
	for _, addr := range p.Addrs {
		if !multiaddr.Contains(addrs, addr) {
			ps.AddAddr(p.ID, addr, time.Hour)
			fmt.Println("Added new peer address to store:", p.ID, addr)
		}
	}

	// Add the peer's addresses to the peer store.
	// for _, addr := range p.Addrs {
	// 	h.Peerstore().AddAddr(p.ID, addr, time.Hour)
	// 	fmt.Println("Added new peer to store:", p.ID, addr)
	// }

	// and connect?
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
