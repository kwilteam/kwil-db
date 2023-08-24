package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/crypto/addresses"

	"github.com/spf13/cobra"
)

func privateKeyCmd() *cobra.Command {
	var keyType, encoding, addressFormat, filePath string
	var overwrite, mute bool

	var cmd = &cobra.Command{
		Use:   "generate-key",
		Short: "Generates a cryptographically secure random private key.",
		Long:  privKeyDesc,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// we set the default address function here because it depends on the key type
			var addressFunc addressCreatorFunc

			var pk crypto.PrivateKey
			var err error
			switch keyType {
			default:
				return cmd.Help()
			case "secp256k1", "": // default
				addressFunc = ethereumAddr
				pk, err = crypto.GenerateSecp256k1Key()
			case "ed25519":
				addressFunc = nearAddr
				pk, err = crypto.GenerateEd25519Key()
			}
			if err != nil {
				return fmt.Errorf("error generating private key: %w", err)
			}

			// check if we need to alter the default address function
			switch addressFormat {
			case "":
				// do nothing
			case "ethereum":
				addressFunc = ethereumAddr
			case "cosmos":
				addressFunc = cosmosAddr
			case "near":
				addressFunc = nearAddr
			default:
				return fmt.Errorf("invalid address format: %s", addressFormat)
			}

			var privKeyStr string
			var pubKeyStr string
			var addrStr string
			switch encoding {
			default:
				return cmd.Help()
			case "hex", "": // default
				privKeyStr = pk.Hex()
				pubKeyStr = hex.EncodeToString(pk.PubKey().Bytes())

				addr, err := addressFunc(pk.PubKey())
				if err != nil {
					return fmt.Errorf("error creating address: %w", err)
				}

				addrStr = hex.EncodeToString(addr.Bytes())
			case "base64":
				privKeyStr = base64.StdEncoding.EncodeToString(pk.Bytes())
				pubKeyStr = base64.StdEncoding.EncodeToString(pk.PubKey().Bytes())

				addr, err := addressFunc(pk.PubKey())
				if err != nil {
					return fmt.Errorf("error creating address: %w", err)
				}
				addrStr = base64.StdEncoding.EncodeToString(addr.Bytes())
			}

			if filePath != "" {
				_, err := os.Stat(filePath)
				if err == nil && !overwrite {
					return fmt.Errorf("file '%s' already exists and overwrite flag is not set", filePath)
				} else if err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("error checking file: %w", err)
				}

				file, err := os.Create(filePath)
				if err != nil {
					return fmt.Errorf("error creating file: %w", err)
				}
				defer file.Close()

				_, err = fmt.Fprintf(file, privKeyStr)
				if err != nil {
					return fmt.Errorf("error writing to file: %w", err)
				}
			}

			if !mute {
				fmt.Printf(printKeyDesc, privKeyStr, pubKeyStr, addrStr)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "secp256k1", "Type of private key to generate: 'secp256k1' or 'ed25519'")
	cmd.Flags().StringVar(&encoding, "encoding", "hex", "Output encoding: 'hex' or 'base64'")
	cmd.Flags().StringVar(&addressFormat, "address-format", "ethereum", "Address format: 'ethereum' or 'cosmos'")
	cmd.Flags().StringVar(&filePath, "file", "", "Write the private key to a file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite the file if it exists")
	cmd.Flags().BoolVar(&mute, "quiet", false, "Mute the output")

	return cmd
}

const privKeyDesc = `The 'generate-key' function generates a cryptographically secure random private key.

This can be used to generate both validator and normal private keys.

To specify the type of private key to generate, pass the '--key-type' flag with either 'secp256k1' or 'ed25519'.
By default, it will generate a secp256k1 private key.

To specify the output encoding, pass the '--encoding <encoding>' flag with either 'hex' or 'base64'.
By default, it will output the private key, public key, and address in hex format.

To specify the outputted address format, pass the '--address-format <format>' flag.
Currently, the CLI supports 'ethereum', 'cosmos', and 'near' address formats.
The default for secp256k1 keys is 'ethereum'.  The default for ed25519 keys is 'near'.

If instead you want to write the private key to a file, pass the '--file <path>' flag.
If a file already exists at the specified path, the operation will fail.
This can be overridden by passing the '--overwrite' flag.

To mute the output, pass the '--quiet' flag.
`

const printKeyDesc = `Private Key: 	%s
Public Key: 	%s
Address: 	%s
`

// addressCreatorFunc is a function that creates an address from a public key.
type addressCreatorFunc func(crypto.PublicKey) (crypto.Address, error)

// nearAddr is an addressCreatorFunc that creates a NEAR address from a public key.
func nearAddr(pk crypto.PublicKey) (crypto.Address, error) {
	edKey, ok := pk.(*crypto.Ed25519PublicKey)
	if !ok {
		return nil, fmt.Errorf("NEAR addresses can only be created from Ed25519 public keys")
	}

	return addresses.CreateNearAddress(edKey)
}

// cosmosAddr is an addressCreatorFunc that creates a Cosmos address from a public key.
func cosmosAddr(pk crypto.PublicKey) (crypto.Address, error) {
	return addresses.CreateCometBFTAddress(pk)
}

// ethereumAddr is an addressCreatorFunc that creates an Ethereum address from a public key.
func ethereumAddr(pk crypto.PublicKey) (crypto.Address, error) {
	secpKey, ok := pk.(*crypto.Secp256k1PublicKey)
	if !ok {
		return nil, fmt.Errorf("ethereum addresses can only be created from Secp256k1 public keys")
	}

	return addresses.CreateEthereumAddress(secpKey)
}
