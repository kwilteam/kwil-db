package key

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/app/shared/display"
	"github.com/kwilteam/kwil-db/core/crypto"

	"github.com/spf13/cobra"
)

var (
	infoLong = `Display information about a private key.

The private key can either be passed as a key file path, or as a hex-encoded string.`

	infoExample = `# Using a key file
kwild key info --key-file ~/.kwild/private_key

# Using a hex-encoded string
kwild key info 381d28cf348c9efbf7d26ea54b647e2cb646d3b98cdeec0f1053a5ff599a036a0aa381bd4aad1670a39977d5416bfac7bd060765adc58a4bb16bbbafeefbae34`
)

func InfoCmd() *cobra.Command {
	var privkeyFile string

	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Display information about a private key.",
		Long:    infoLong,
		Example: infoExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// if len(args) == 1, then the private key is passed as a hex string
			// otherwise, it is passed as a file path
			if len(args) == 1 {
				key, err := hex.DecodeString(args[0])
				if err != nil {
					return display.PrintErr(cmd, fmt.Errorf("private key not valid hex: %w", err))
				}
				return display.PrintCmd(cmd, privKeyInfo(key, crypto.KeyTypeSecp256k1))
			} else if privkeyFile != "" {
				key, err := readKeyFile(privkeyFile)
				if err != nil {
					return display.PrintErr(cmd, err)
				}
				return display.PrintCmd(cmd, privKeyInfo(key, crypto.KeyTypeSecp256k1))
			}

			cmd.Usage()
			return errors.New("must provide with the private key file or hex string")
		},
	}

	cmd.Flags().StringVarP(&privkeyFile, "key-file", "o", "", "file containing the private key to display")

	return cmd
}

// readKeyFile reads a private key from a text file containing the hexadecimal
// encoding of the private key bytes.
func readKeyFile(keyFile string) ([]byte, error) {
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
