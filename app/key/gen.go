package key

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/core/crypto"

	"github.com/spf13/cobra"
)

var (
	genExample = `# Generate a new key and save it to ./priv_key
kwild key gen --key-file ./priv_key

# Generate a raw private key
kwild key gen --raw`
)

func GenCmd() *cobra.Command {
	var raw bool // if true, output hex private key only
	var out string

	cmd := &cobra.Command{
		Use:     "gen",
		Short:   "Generate private keys for usage in validators.",
		Long:    "Generate private keys for usage in validators.",
		Example: genExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := generatePrivateKey()
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if out == "" {
				if raw {
					return display.PrintCmd(cmd, display.RespString(hex.EncodeToString(privKey.Bytes())))
				} else {
					pki := &PrivateKeyInfo{
						PrivateKeyHex: hex.EncodeToString(privKey.Bytes()),
						PublicKeyHex:  hex.EncodeToString(privKey.Public().Bytes()),
					}
					return display.PrintCmd(cmd, pki)
				}
			}

			err = os.WriteFile(out, []byte(hex.EncodeToString(privKey.Bytes())), 0600)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespString("Private key written to "+out))
		},
	}

	cmd.Flags().BoolVarP(&raw, "raw", "R", false, "just print the private key hex without other encodings, public key, or node ID")
	cmd.Flags().StringVarP(&out, "key-file", "o", "", "file to which the new private key is written (stdout by default)")

	return cmd
}

func generatePrivateKey( /* TODO: key type */ ) (crypto.PrivateKey, error) {
	privKey, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	return privKey, err
}
