package key

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/spf13/cobra"
)

const keyExplain = "The `key` command provides subcommands for private key generation and inspection."

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: keyExplain,
	Long:  "The `key` command provides subcommands for private key generation and inspection. These are the private keys that identify the node on the network and provide validator transaction signing capability.",
}

func NewKeyCmd() *cobra.Command {
	cmd := keyCmd

	// Add subcommands
	cmd.AddCommand(
		genCmd(),
		infoCmd(),
	)

	return cmd
}

func privKeyInfo(privateKey []byte) *PrivateKeyInfo {
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
	type pki PrivateKeyInfo
	return json.Marshal((*pki)(p))
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
