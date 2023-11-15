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

func PrivKeyInfo(privateKey []byte) *PrivateKeyInfo {
	priv := ed25519.PrivKey(privateKey)
	pub := priv.PubKey().(ed25519.PubKey)
	nodeID := p2p.PubKeyToID(pub)

	return &PrivateKeyInfo{
		PrivateKeyHex:         hex.EncodeToString(priv.Bytes()),
		PrivateKeyBase64:      base64.StdEncoding.EncodeToString(priv.Bytes()),
		PublicKeyBase64:       base64.StdEncoding.EncodeToString(pub.Bytes()),
		PublicKeyCometizedHex: pub.String(),
		PublicKeyPlainHex:     hex.EncodeToString(pub.Bytes()),
		Address:               pub.Address().String(),
		NodeID:                fmt.Sprintf("%v", nodeID), // same as address, just upper case
	}
}

type PrivateKeyInfo struct {
	PrivateKeyHex         string `json:"private_key_hex"`
	PrivateKeyBase64      string `json:"private_key_base64"`
	PublicKeyBase64       string `json:"public_key_base64"`
	PublicKeyCometizedHex string `json:"public_key_cometized_hex"`
	PublicKeyPlainHex     string `json:"public_key_plain_hex"`
	Address               string `json:"address"`
	NodeID                string `json:"node_id"`
}

func (p *PrivateKeyInfo) MarshalJSON() ([]byte, error) {
	// must use anonymous struct to avoid infinite recursion
	return json.Marshal(struct {
		PrivateKeyHex         string `json:"private_key_hex"`
		PrivateKeyBase64      string `json:"private_key_base64"`
		PublicKeyBase64       string `json:"public_key_base64"`
		PublicKeyCometizedHex string `json:"public_key_cometized_hex"`
		PublicKeyPlainHex     string `json:"public_key_plain_hex"`
		Address               string `json:"address"`
		NodeID                string `json:"node_id"`
	}{
		PrivateKeyHex:         p.PrivateKeyHex,
		PrivateKeyBase64:      p.PrivateKeyBase64,
		PublicKeyBase64:       p.PublicKeyBase64,
		PublicKeyCometizedHex: p.PublicKeyCometizedHex,
		PublicKeyPlainHex:     p.PublicKeyPlainHex,
		Address:               p.Address,
		NodeID:                p.NodeID,
	})
}

func (p *PrivateKeyInfo) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(`Private key (hex): %s
Private key (base64): %s
Public key (base64): %s
Public key (cometized hex): %v
Public key (plain hex): %v
Address (string): %s
Node ID: %v`,
		p.PrivateKeyHex,
		p.PrivateKeyBase64,
		p.PublicKeyBase64,
		p.PublicKeyCometizedHex,
		p.PublicKeyPlainHex,
		p.Address,
		p.NodeID,
	)), nil
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
