package key

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/shared/display"
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
	display.BindOutputFormatFlag(keyCmd)
	return keyCmd
}

func privKeyInfo(priv crypto.PrivateKey) *PrivateKeyInfo {
	return &PrivateKeyInfo{
		KeyType:       priv.Type().String(),
		PrivateKeyHex: hex.EncodeToString(priv.Bytes()),
		PublicKeyHex:  hex.EncodeToString(priv.Public().Bytes()),
	}
}

type PrivateKeyInfo struct {
	KeyType       string `json:"key_type"`
	PrivateKeyHex string `json:"private_key_hex"`
	PublicKeyHex  string `json:"public_key_hex"`
}

func (p *PrivateKeyInfo) MarshalJSON() ([]byte, error) {
	type pki PrivateKeyInfo
	return json.Marshal((*pki)(p))
}

func (p *PrivateKeyInfo) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(`Key type: %s
Private key (hex): %s
Public key (plain hex): %v`,
		p.KeyType,
		p.PrivateKeyHex,
		p.PublicKeyHex,
	)), nil
}
