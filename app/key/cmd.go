package key

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/core/crypto"
)

const keyExplain = "The `key` command provides subcommands for private key generation and inspection."

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: keyExplain,
	Long:  "The `key` command provides subcommands for private key generation and inspection. These are the private keys that identify the node on the network and provide validator transaction signing capability.",
}

func KeyCmd() *cobra.Command {
	keyCmd.AddCommand(
		GenCmd(),
		InfoCmd(),
	)
	return keyCmd
}

func privKeyInfo(privateKey []byte, keyType crypto.KeyType) *PrivateKeyInfo {
	priv, err := crypto.UnmarshalPrivateKey(privateKey, keyType)
	if err != nil {
		return &PrivateKeyInfo{PrivateKeyHex: "<invalid>"}
	}
	pub := priv.Public()

	return &PrivateKeyInfo{
		PrivateKeyHex: hex.EncodeToString(priv.Bytes()),
		PublicKeyHex:  hex.EncodeToString(pub.Bytes()),
	}
}

type PrivateKeyInfo struct {
	PrivateKeyHex string `json:"private_key_hex"`
	PublicKeyHex  string `json:"public_key_hex"`
}

func (p *PrivateKeyInfo) MarshalJSON() ([]byte, error) {
	type pki PrivateKeyInfo
	return json.Marshal((*pki)(p))
}

func (p *PrivateKeyInfo) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(`Private key (hex): %s
Public key (plain hex): %v`,
		p.PrivateKeyHex,
		p.PublicKeyHex,
	)), nil
}
