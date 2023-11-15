package key

import (
	"encoding/hex"
	"os"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/internal/abci"
	"github.com/spf13/cobra"
)

var (
	genExample = `# Generate a new key and save it to ./priv_key
kwil-admin key gen --key-file ./priv_key

# Generate a raw private key
kwil-admin key gen --raw`
)

func genCmd() *cobra.Command {
	var raw bool // if true, output hex private key only
	var out string

	cmd := &cobra.Command{
		Use:     "gen",
		Short:   "Generate ed25519 keys for usage in validators.",
		Long:    "Generate ed25519 keys for usage in validators.",
		Example: genExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey := abci.GeneratePrivateKey()
			if out == "" {
				if raw {
					return display.PrintCmd(cmd, display.RespString(hex.EncodeToString(privKey)))
				} else {
					return display.PrintCmd(cmd, abci.PrivKeyInfo(privKey))
				}
			}

			return os.WriteFile(out, []byte(hex.EncodeToString(privKey[:])), 0600)
		},
	}

	cmd.Flags().BoolVarP(&raw, "raw", "R", false, "just print the private key hex without other encodings, public key, or node ID")
	cmd.Flags().StringVarP(&out, "key-file", "o", "", "file to which the new private key is written (stdout by default)")

	return cmd
}
