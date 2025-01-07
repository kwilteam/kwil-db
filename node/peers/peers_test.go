package peers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestPersistAndLoadPeers(t *testing.T) {
	ma1, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/4001")
	ma1b, _ := ma.NewMultiaddr("/ip4/127.0.0.2/tcp/4001")
	pid1, _ := peer.Decode("16Uiu2HAm8iRUsTzYepLP8pdJL3645ACP7VBfZQ7yFbLfdb7WvkL7")
	pk1, err := pubKeyFromPeerID(pid1)
	if err != nil {
		t.Fatal(err)
	}
	nid1 := NodeIDFromPubKey(pk1)
	ma2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/4002")
	pid2, _ := peer.Decode("16Uiu2HAkx2kfP117VnYnaQGprgXBoMpjfxGXCpizju3cX7ZUzRhv")
	pk2, err := pubKeyFromPeerID(pid2)
	if err != nil {
		t.Fatal(err)
	}
	nid2 := NodeIDFromPubKey(pk2)

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_peers.json")

	testPeers := []PersistentPeerInfo{
		{
			NodeID:      nid1,
			Addrs:       []ma.Multiaddr{ma1, ma1b},
			Protos:      []protocol.ID{"ProtocolWhatever"},
			Whitelisted: true,
		},
		{
			NodeID:      nid2,
			Addrs:       []ma.Multiaddr{ma2},
			Protos:      []protocol.ID{"ProtocolWhatever", "ProtocolOther"},
			Whitelisted: false,
		},
	}

	t.Run("persist and load peers successfully", func(t *testing.T) {
		err := persistPeers(testPeers, testFile)
		require.NoError(t, err)

		t.Log(testFile)

		loadedPeers, err := loadPeers(testFile)
		require.NoError(t, err)
		require.Equal(t, testPeers, loadedPeers)
	})

	t.Run("persist empty peer list", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty_peers.json")
		err := persistPeers([]PersistentPeerInfo{}, emptyFile)
		require.NoError(t, err)

		loadedPeers, err := loadPeers(emptyFile)
		require.NoError(t, err)
		require.Empty(t, loadedPeers)
	})

	t.Run("persist empty peer list file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty_peers_file.json")
		fid, err := os.Create(emptyFile)
		require.NoError(t, err)
		fid.Close()

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
