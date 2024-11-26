package node

import (
	"context"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

// This file has some building blocks for DHT that can be used for both content
// routing and peer discovery.
//
// One tentative approach for snapshots is:
//  1. discover snapshot providers (peer discover)
//  2. retrieve their snapshot catalogs (ProtocolIDSnapshotCatalog), which describe their provided snapshots (height, hash, chunk count)
//  3. aggregate the catalogs and pick the best snapshot
//  4. begin retrieving the chunks (ProtocolIDSnapshotChunk)
//
// Higher level logic will be needed for the aggregation, and fallback to
// next-best shapshots in the event that restore of the current best fails.

const (
	snapshotCatalogNS     = "snapshot-catalog"
	snapshotChunkNSPrefix = "snapshot-chunk/" // e.g. "snapshot-chunk/{blockHash}/{chunkIdx}"
)

func makeDHT(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	// Create a DHT
	kadDHT, err := dht.New(ctx, h /*, dht.BootstrapPeers()*/)
	if err != nil {
		return nil, err
	}

	// Bootstrap DHT
	err = kadDHT.Bootstrap(ctx)
	if err != nil {
		kadDHT.Close()
		return nil, err
	}

	return kadDHT, nil
}

func makeDiscovery(kad *dht.IpfsDHT) discovery.Discovery {
	return drouting.NewRoutingDiscovery(kad)
}

func provide(ctx context.Context, namespace string, a discovery.Advertiser) error {
	_ /*ttl*/, err := a.Advertise(ctx, namespace /*, discovery.TTL(25*time.Hour)*/)
	// now caller should handle requests from peers that discover this content:
	//  h.SetStreamHandler(SomeProtocolID, func(s network.Stream) {
	return err
}

func discoverProviders(ctx context.Context, namespace string, limit int, d discovery.Discoverer) ([]peer.AddrInfo, error) {
	peerChan, err := d.FindPeers(ctx, namespace, discovery.Limit(limit))
	if err != nil {
		return nil, err
	}
	var peers []peer.AddrInfo
	for p := range peerChan {
		peers = append(peers, p)
	}
	return peers, nil
	// now caller may open a stream to the providing peer(s) with the appropriate protocol:
	// h.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.TempAddrTTL)
	// stream, err := h.NewStream(ctx, peer.ID, SomeProtocolID)
}
