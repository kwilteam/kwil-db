package specifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func AddPeerSpecification(ctx context.Context, t *testing.T, netops PeersDsl, peerID string) {
	t.Log("Executing add peer specification")

	peers, err := netops.ListPeers(ctx)
	require.NoError(t, err)
	require.False(t, hasPeer(peers, peerID))

	err = netops.AddPeer(ctx, peerID)
	require.NoError(t, err)

	peers, err = netops.ListPeers(ctx)
	require.NoError(t, err)
	require.True(t, hasPeer(peers, peerID))

	t.Logf("Added peer %s", peerID)
}

func AddExistingPeerSpecification(ctx context.Context, t *testing.T, netops PeersDsl, peerID string) {
	t.Log("Executing add existing peer specification")

	peers, err := netops.ListPeers(ctx)
	require.NoError(t, err)
	require.True(t, hasPeer(peers, peerID))

	err = netops.AddPeer(ctx, peerID)
	require.Error(t, err)

	peers, err = netops.ListPeers(ctx)
	require.NoError(t, err)
	require.True(t, hasPeer(peers, peerID))
}

func RemovePeerSpecification(ctx context.Context, t *testing.T, netops PeersDsl, peerID string) {
	t.Log("Executing remove peer specification")

	peers, err := netops.ListPeers(ctx)
	require.NoError(t, err)
	require.True(t, hasPeer(peers, peerID))

	err = netops.RemovePeer(ctx, peerID)
	require.NoError(t, err)

	peers, err = netops.ListPeers(ctx)
	require.NoError(t, err)
	require.False(t, hasPeer(peers, peerID))

	t.Logf("Removed peer %s", peerID)
}

func RemoveNonExistingPeerSpecification(ctx context.Context, t *testing.T, netops PeersDsl, peerID string) {
	t.Log("Executing remove non-existing peer specification")

	peers, err := netops.ListPeers(ctx)
	require.NoError(t, err)
	require.False(t, hasPeer(peers, peerID))

	err = netops.RemovePeer(ctx, peerID)
	require.Error(t, err)

	peers, err = netops.ListPeers(ctx)
	require.NoError(t, err)
	require.False(t, hasPeer(peers, peerID))
}

func ListPeersSpecification(ctx context.Context, t *testing.T, netops PeersDsl, peers []string) {
	t.Log("Executing list peers specification")

	peersList, err := netops.ListPeers(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, peers, peersList)
}

func hasPeer(peers []string, peerID string) bool {
	for _, peer := range peers {
		if peer == peerID {
			return true
		}
	}
	return false
}

func PeerConnectivitySpecification(ctx context.Context, t *testing.T, netops PeersDsl, peerID string, connected bool) {
	t.Log("Executing peer connectivity specification")

	connectedPeers, err := netops.ConnectedPeers(ctx)
	require.NoError(t, err)

	isConnected := hasPeer(connectedPeers, peerID)
	if connected {
		require.True(t, isConnected)
	} else {
		require.False(t, isConnected)
	}
}
