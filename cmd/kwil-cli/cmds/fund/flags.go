package fund

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/crypto"

	"github.com/spf13/cobra"
)

const (
	addressFlag = "address"
)

// getAddress returns the address flag value.
// If no value is passed, it will use the address of the user's wallet.
func getSelectedAddress(cmd *cobra.Command, conf *config.KwilCliConfig) (string, error) {
	if cmd.Flags().Changed(addressFlag) {
		return cmd.Flags().GetString(addressFlag)
	}

	return crypto.AddressFromPrivateKey(conf.PrivateKey), nil
}
