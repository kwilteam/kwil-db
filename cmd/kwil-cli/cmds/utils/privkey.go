package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/crypto/addresses"
	"github.com/spf13/cobra"
)

// generatePrivateKey is a helper function that generates a private key info
func generatePrivateKey(keyType string, addressFormat string, encoding string) (*generatedWalletInfo, error) {
	var pk crypto.PrivateKey
	var err error
	var addressFunc addressCreatorFunc

	generatedKeyInfo := &generatedWalletInfo{}

	keyVariant := fmt.Sprintf("%s %s", keyType, addressFormat)

	switch keyVariant {
	case "secp256k1 ethereum":
		addressFunc = ethereumAddr
		pk, err = crypto.GenerateSecp256k1Key()
	case "secp256k1 cometbft":
		addressFunc = cometbftAddr
		pk, err = crypto.GenerateSecp256k1Key()
	case "ed25519 near":
		addressFunc = nearAddr
		pk, err = crypto.GenerateEd25519Key()
	default:
		return nil, fmt.Errorf(
			"not supported combination: %s", keyVariant)
	}

	if err != nil {
		return nil, fmt.Errorf("error generating private key: %w", err)
	}

	addr, err := addressFunc(pk.PubKey())
	if err != nil {
		return nil, fmt.Errorf("error derive address: %w", err)
	}

	generatedKeyInfo.Address = addr.String()

	switch encoding {
	case "hex", "": // default
		generatedKeyInfo.PrivateKey = pk.Hex()
		generatedKeyInfo.PublicKey = hex.EncodeToString(pk.PubKey().Bytes())
	case "base64":
		generatedKeyInfo.PrivateKey = base64.StdEncoding.EncodeToString(pk.Bytes())
		generatedKeyInfo.PublicKey = base64.StdEncoding.EncodeToString(pk.PubKey().Bytes())
	default:
		return nil, fmt.Errorf("not supported encoding: %s", encoding)
	}

	return generatedKeyInfo, nil
}

func privateKeyCmd() *cobra.Command {
	var keyType, encoding, addressFormat, filePath string
	var overwrite, mute bool

	var cmd = &cobra.Command{
		Use:   "generate-key",
		Short: "Generates a cryptographically secure random private key.",
		Long:  privKeyDesc,
		RunE: func(cmd *cobra.Command, _ []string) error {
			respGenKeyInfo := &respGenWalletInfo{}

			err := func() error {
				generatedKeyInfo, err := generatePrivateKey(keyType, addressFormat, encoding)
				if err != nil {
					return err
				}

				if filePath != "" {
					_, err := os.Stat(filePath)
					if err == nil && !overwrite {
						return fmt.Errorf("file '%s' already exists and overwrite flag is not set", filePath)
					} else if err != nil && !os.IsNotExist(err) {
						return fmt.Errorf("error checking file: %w", err)
					}

					err = os.WriteFile(filePath, []byte(generatedKeyInfo.PrivateKey), 0600)
					if err != nil {
						return fmt.Errorf("error writing to file: %w", err)
					}
				}

				respGenKeyInfo.info = generatedKeyInfo
				return nil
			}()

			if mute {
				return err
			}

			return display.Print(respGenKeyInfo, err, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "ethereum", "Type of private key to generate: 'secp256k1' or 'ed25519'")
	cmd.Flags().StringVar(&encoding, "encoding", "hex", "Output encoding: 'hex' or 'base64'")
	cmd.Flags().StringVar(&addressFormat, "address-format", "ethereum", "Address format: 'ethereum' or 'cometbft' or 'near'")
	cmd.Flags().StringVar(&filePath, "file", "", "Write the private key to a file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite the file if it exists")
	cmd.Flags().BoolVar(&mute, "quiet", false, "Mute the output")

	return cmd
}

const privKeyDesc = `The 'generate-key' function generates a cryptographically secure random private key.

This can be used to generate both validator and normal private keys.

To specify the type of private key to generate, pass the '--key-type' flag with either 'secp256k1' or 'ed25519'.
By default, it will generate a secp256k1 private key.

To specify the outputted address format, pass the '--address-format <format>' flag.
Currently, the CLI supports 'ethereum', 'cometbft', and 'near' address formats.
The default for secp256k1 keys is 'ethereum'.  The default for ed25519 keys is 'near'.

To specify the output encoding, pass the '--encoding <encoding>' flag with either 'hex' or 'base64'.
By default, it will output the private keyand public key in hex format.  The '--encoding' flag
only affects the private key and public key, not the address.  The address will always be outputted
as the canonical string representation of the address type.

If instead you want to write the private key to a file, pass the '--file <path>' flag.
If a file already exists at the specified path, the operation will fail.
This can be overridden by passing the '--overwrite' flag.

To mute the output, pass the '--quiet' flag.
`

// addressCreatorFunc is a function that creates an address from a public key.
type addressCreatorFunc func(crypto.PublicKey) (crypto.Address, error)

// nearAddr is an addressCreatorFunc that creates a NEAR address from a public key.
func nearAddr(pk crypto.PublicKey) (crypto.Address, error) {
	return addresses.GenerateAddress(pk, addresses.AddressFormatNEAR)
}

// cometbftAddr is an addressCreatorFunc that creates a cometbft address from a public key.
func cometbftAddr(pk crypto.PublicKey) (crypto.Address, error) {
	return addresses.GenerateAddress(pk, addresses.AddressFormatCometBFT)
}

// ethereumAddr is an addressCreatorFunc that creates an Ethereum address from a public key.
func ethereumAddr(pk crypto.PublicKey) (crypto.Address, error) {
	return addresses.GenerateAddress(pk, addresses.AddressFormatEthereum)
}
