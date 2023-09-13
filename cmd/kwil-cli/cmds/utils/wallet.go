package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/spf13/cobra"
)

// generateWallet is a helper function that generates a wallet
func generateWallet(walletType string, encoding string) (*generatedWalletInfo, error) {
	var pk crypto.PrivateKey
	var err error
	var addressFunc addressCreatorFunc

	generatedKeyInfo := &generatedWalletInfo{}

	switch walletType {
	case "ethereum":
		addressFunc = ethereumAddr
		pk, err = crypto.GenerateSecp256k1Key()
	case "cometbft":
		addressFunc = cometbftAddr
		pk, err = crypto.GenerateSecp256k1Key()
	case "near":
		addressFunc = nearAddr
		pk, err = crypto.GenerateEd25519Key()
	default:
		return nil, fmt.Errorf(
			"not supported combination: %s", walletType)
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

func walletCmd() *cobra.Command {
	var walletType, encoding, filePath string
	var overwrite, mute bool

	var cmd = &cobra.Command{
		Use:   "generate-wallet",
		Short: "Generates a wallet.",
		Long:  walletDesc,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var resp respGenWalletInfo

			err := func() error {
				generatedKeyInfo, err := generateWallet(walletType, encoding)
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

				resp.info = generatedKeyInfo
				return nil
			}()

			if mute {
				return err
			}

			return display.Print(&resp, err, config.GetOutputFormat())
		},
	}

	cmd.Flags().StringVar(&walletType, "wallet-type", "ethereum", "Type of wallet: 'ethereum' or 'cometbft' or 'near'")
	cmd.Flags().StringVar(&encoding, "encoding", "hex", "Output encoding: 'hex' or 'base64'")
	cmd.Flags().StringVar(&filePath, "file", "", "Write the private key to a file")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite the file if it exists")
	cmd.Flags().BoolVar(&mute, "quiet", false, "Mute the output")

	return cmd
}

const walletDesc = `The 'generate-wallet' function generates a wallet.

This can be used to generate both validator and normal private keys.

To specify the type of wallet to generate, pass the '--wallet-type' flag with either 'ethereum' or 'comebft' or 'near'".
By default, it will generate a ethereum wallet.

To specify the output encoding, pass the '--encoding <encoding>' flag with either 'hex' or 'base64'.
By default, it will output the private key and public key in hex format.  The '--encoding' flag
only affects the private key and public key, not the address.  The address will always be outputted
as the canonical string representation of the address type.

If instead you want to write the private key to a file, pass the '--file <path>' flag.
If a file already exists at the specified path, the operation will fail.
This can be overridden by passing the '--overwrite' flag.

To mute the output, pass the '--quiet' flag.
`
