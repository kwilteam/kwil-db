package key

import (
	"errors"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/internal/abci"
	"github.com/spf13/cobra"
)

var (
	infoLong = `Display information about a private key.

The private key can either be passed as a key file path, or as a hex-encoded string.`

	infoExample = `# Using a key file
kwil-admin key info --key-file ~/.kwild/private_key

# Using a hex-encoded string
kwil-admin key info 381d28cf348c9efbf7d26ea54b647e2cb646d3b98cdeec0f1053a5ff599a036a0aa381bd4aad1670a39977d5416bfac7bd060765adc58a4bb16bbbafeefbae34`
)

func infoCmd() *cobra.Command {
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
				return display.PrintCmd(cmd, abci.PrivKeyInfo([]byte(args[0])))
			} else if privkeyFile != "" {
				key, err := abci.ReadKeyFile(privkeyFile)
				if err != nil {
					return err
				}

				return display.PrintCmd(cmd, abci.PrivKeyInfo(key))
			} else {
				return errors.New("must provide with the private key file or hex string")
			}
		},
	}

	cmd.Flags().StringVarP(&privkeyFile, "key-file", "o", "", "file containing the private key to display")

	return cmd
}
