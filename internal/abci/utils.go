package abci

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/internal/abci/snapshots"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
)

// cometAddrFromPubKey computes the cometBFT address from an ed25519 public key.
// This helper is required to support the application's BeginBlock method where
// the RequestBeginBlock request type includes the addresses of any byzantine
// validators rather than their public keys, as with all of the other methods.
func cometAddrFromPubKey(pubkey []byte) string {
	publicKey := ed25519.PubKey(pubkey)
	return publicKey.Address().String()
}

func convertABCISnapshots(req *abciTypes.Snapshot) *snapshots.Snapshot {
	var metadata snapshots.SnapshotMetadata
	err := json.Unmarshal(req.Metadata, &metadata)
	if err != nil {
		return nil
	}

	snapshot := &snapshots.Snapshot{
		Height:     req.Height,
		Format:     req.Format,
		ChunkCount: req.Chunks,
		Hash:       req.Hash,
		Metadata:   metadata,
	}
	return snapshot
}

func convertToABCISnapshot(snapshot *snapshots.Snapshot) (*abciTypes.Snapshot, error) {
	metadata, err := json.Marshal(snapshot.Metadata)
	if err != nil {
		return nil, err
	}

	return &abciTypes.Snapshot{
		Height:   snapshot.Height,
		Format:   snapshot.Format,
		Chunks:   snapshot.ChunkCount,
		Hash:     snapshot.Hash,
		Metadata: metadata,
	}, nil
}

func abciStatus(status snapshots.Status) abciTypes.ResponseApplySnapshotChunk_Result {
	switch status {
	case snapshots.ACCEPT:
		return abciTypes.ResponseApplySnapshotChunk_ACCEPT
	case snapshots.REJECT:
		return abciTypes.ResponseApplySnapshotChunk_REJECT_SNAPSHOT
	case snapshots.RETRY:
		return abciTypes.ResponseApplySnapshotChunk_RETRY
	default:
		return abciTypes.ResponseApplySnapshotChunk_UNKNOWN
	}
}

func PrintPrivKeyInfo(privateKey []byte) {
	priv := ed25519.PrivKey(privateKey)
	pub := priv.PubKey().(ed25519.PubKey)
	nodeID := p2p.PubKeyToID(pub)

	fmt.Printf("Private key (hex): %s\n", hex.EncodeToString(priv.Bytes()))
	fmt.Printf("Private key (base64): %s\n",
		base64.StdEncoding.EncodeToString(priv.Bytes())) // "value" in abci/config/node_key.json
	fmt.Printf("Public key (base64): %s\n",
		base64.StdEncoding.EncodeToString(pub.Bytes())) // "validators.pub_key.value" in abci/config/genesis.json
	fmt.Printf("Public key (cometized hex): %v\n", pub.String())                // for reference with some cometbft logs
	fmt.Printf("Public key (plain hex): %v\n", hex.EncodeToString(pub.Bytes())) // for reference with some cometbft logs
	fmt.Printf("Address (string): %s\n", pub.Address().String())                // "validators.address" in abci/config/genesis.json
	fmt.Printf("Node ID: %v\n", nodeID)                                         // same as address, just upper case
}

func GeneratePrivateKey() []byte {
	privKey := ed25519.GenPrivKey()
	return privKey[:]
}

// ReadKeyFile reads a private key from a text file containing the hexadecimal
// encoding of the private key bytes.
func ReadKeyFile(keyFile string) ([]byte, error) {
	privKeyHexB, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %v", err)
	}
	privKeyHex := string(bytes.TrimSpace(privKeyHexB))
	privB, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("error decoding private key: %v", err)
	}
	return privB, nil
}
