package key

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/core/crypto"
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
		Use:     "gen [<keytype>]",
		Short:   "Generate private keys for usage in validators.",
		Long:    "Generate private keys for usage in validators.",
		Example: genExample,
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyType := crypto.KeyTypeSecp256k1 // default with 0 args
			if len(args) > 0 {
				var err error
				keyType, err = crypto.ParseKeyType(args[0])
				if err != nil {
					keyTypeInt, err := strconv.ParseUint(args[0], 10, 16)
					if err != nil {
						return display.PrintErr(cmd, fmt.Errorf("invalid key type (%s): %w", args[0], err))
					}
					keyType = crypto.KeyType(keyTypeInt)
				}
			}

			privKey, err := crypto.GeneratePrivateKey(keyType)
			if err != nil {
				return display.PrintErr(cmd, err)
			}

			if out == "" {
				if raw {
					return display.PrintCmd(cmd, display.RespString(hex.EncodeToString(privKey.Bytes())))
				}
				return display.PrintCmd(cmd, &PrivateKeyInfo{
					KeyType:       keyType.String(),
					PrivateKeyHex: hex.EncodeToString(privKey.Bytes()),
					PublicKeyHex:  hex.EncodeToString(privKey.Public().Bytes()),
				})
			}

			if err := SaveNodeKey(out, privKey); err != nil {
				return display.PrintErr(cmd, err)
			}

			return display.PrintCmd(cmd, display.RespString("Private key written to "+out))
		},
	}

	cmd.Flags().BoolVarP(&raw, "raw", "R", false, "just print the private key hex without other encodings, public key, or node ID")
	cmd.Flags().StringVarP(&out, "key-file", "o", "", "file to which the new private key is written (stdout by default)")

	return cmd
}
