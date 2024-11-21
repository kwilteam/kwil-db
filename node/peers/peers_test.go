package peers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/node/types"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestPersistAndLoadPeers(t *testing.T) {
	ma1, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/4001")
	ma1b, _ := ma.NewMultiaddr("/ip4/127.0.0.2/tcp/4001")
	pid1, _ := peer.Decode("16Uiu2HAm8iRUsTzYepLP8pdJL3645ACP7VBfZQ7yFbLfdb7WvkL7")
	ma2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/4002")
	pid2, _ := peer.Decode("16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv")

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_peers.json")

	testPeers := []types.PeerInfo{
		{
			AddrInfo: types.AddrInfo{
				ID:    pid1,
				Addrs: []ma.Multiaddr{ma1, ma1b},
			},
			Protos: []protocol.ID{"ProtocolWhatever"},
		},
		{
			AddrInfo: types.AddrInfo{
				ID:    pid2,
				Addrs: []ma.Multiaddr{ma2},
			},
			Protos: []protocol.ID{"ProtocolWhatever", "ProtocolOther"},
		},
	}

	t.Run("persist and load peers successfully", func(t *testing.T) {
		err := persistPeers(testPeers, testFile)
		require.NoError(t, err)

		loadedPeers, err := loadPeers(testFile)
		require.NoError(t, err)
		require.Equal(t, testPeers, loadedPeers)
	})

	t.Run("persist empty peer list", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty_peers.json")
		err := persistPeers([]types.PeerInfo{}, emptyFile)
		require.NoError(t, err)

		loadedPeers, err := loadPeers(emptyFile)
		require.NoError(t, err)
		require.Empty(t, loadedPeers)
	})

	t.Run("load from non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "non_existent.json")
		_, err := loadPeers(nonExistentFile)
		require.Error(t, err)
	})

	t.Run("load from invalid JSON file", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(invalidFile, []byte("invalid json"), 0644)
		require.NoError(t, err)

		_, err = loadPeers(invalidFile)
		require.Error(t, err)
	})

	t.Run("persist to read-only directory", func(t *testing.T) {
		readOnlyDir := filepath.Join(tempDir, "readonly")
		require.NoError(t, os.Mkdir(readOnlyDir, 0444))
		readOnlyFile := filepath.Join(readOnlyDir, "peers.json")

		err := persistPeers(testPeers, readOnlyFile)
		require.Error(t, err)
	})
}
