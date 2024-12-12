package peers

import (
	"encoding/json"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestPeerInfoJSON(t *testing.T) {
	addr1, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	addr2, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/5678")
	pid, _ := peer.Decode("16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv")

	tests := []struct {
		name     string
		peerInfo PeerInfo
	}{
		{
			name: "basic peer info",
			peerInfo: PeerInfo{
				AddrInfo: AddrInfo{
					ID:    pid,
					Addrs: []multiaddr.Multiaddr{addr1, addr2},
				},
				Protos: []protocol.ID{"/proto/1.0.0", "/proto/2.0.0"},
			},
		},
		{
			name: "empty addresses",
			peerInfo: PeerInfo{
				AddrInfo: AddrInfo{
					ID:    pid,
					Addrs: []multiaddr.Multiaddr{},
				},
				Protos: []protocol.ID{"/proto/1.0.0"},
			},
		},
		{
			name: "empty protocols",
			peerInfo: PeerInfo{
				AddrInfo: AddrInfo{
					ID:    pid,
					Addrs: []multiaddr.Multiaddr{addr1},
				},
				Protos: []protocol.ID{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.peerInfo)
			require.NoError(t, err)

			t.Log(string(data))

			var decoded PeerInfo
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			require.Equal(t, tt.peerInfo.ID, decoded.ID)
			require.Equal(t, len(tt.peerInfo.Addrs), len(decoded.Addrs))
			require.Equal(t, len(tt.peerInfo.Protos), len(decoded.Protos))

			for i, addr := range tt.peerInfo.Addrs {
				require.Equal(t, addr.String(), decoded.Addrs[i].String())
			}
			for i, proto := range tt.peerInfo.Protos {
				require.Equal(t, proto, decoded.Protos[i])
			}
		})
	}
}

func TestPersistentPeerInfoJSON(t *testing.T) {
	addr1, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	pid, _ := peer.Decode("16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv")
	pk, err := pubKeyFromPeerID(pid)
	if err != nil {
		t.Fatal(err)
	}
	nid := NodeIDFromPubKey(pk)

	t.Log(nid)

	tests := []struct {
		name     string
		peerInfo PersistentPeerInfo
	}{
		{
			name: "whitelisted peer",
			peerInfo: PersistentPeerInfo{
				NodeID:      nid,
				Addrs:       []multiaddr.Multiaddr{addr1},
				Protos:      []protocol.ID{"/proto/1.0.0"},
				Whitelisted: true,
			},
		},
		{
			name: "non-whitelisted peer",
			peerInfo: PersistentPeerInfo{
				NodeID:      nid,
				Addrs:       []multiaddr.Multiaddr{addr1},
				Protos:      []protocol.ID{"/proto/1.0.0"},
				Whitelisted: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.peerInfo)
			require.NoError(t, err)

			t.Log(string(data))

			var decoded PersistentPeerInfo
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			require.Equal(t, tt.peerInfo.NodeID, decoded.NodeID)
			require.Equal(t, tt.peerInfo.Whitelisted, decoded.Whitelisted)
			require.Equal(t, len(tt.peerInfo.Addrs), len(decoded.Addrs))
			require.Equal(t, len(tt.peerInfo.Protos), len(decoded.Protos))

			for i, addr := range tt.peerInfo.Addrs {
				require.Equal(t, addr.String(), decoded.Addrs[i].String())
			}
			for i, proto := range tt.peerInfo.Protos {
				require.Equal(t, proto, decoded.Protos[i])
			}
		})
	}
}
